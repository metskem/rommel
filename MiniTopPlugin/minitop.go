package main

import (
	"code.cloudfoundry.org/cli/plugin"
	"code.cloudfoundry.org/go-loggregator/v9"
	"code.cloudfoundry.org/go-loggregator/v9/rpc/loggregator_v2"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/awesome-gocui/gocui"
	"github.com/integrii/flaggy"
	"github.com/metskem/rommel/MiniTopPlugin/apps"
	"github.com/metskem/rommel/MiniTopPlugin/common"
	"github.com/metskem/rommel/MiniTopPlugin/conf"
	"github.com/metskem/rommel/MiniTopPlugin/util"
	"github.com/metskem/rommel/MiniTopPlugin/vms"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	allSelectors = []*loggregator_v2.Selector{
		{Message: &loggregator_v2.Selector_Gauge{Gauge: &loggregator_v2.GaugeSelector{}}},
		{Message: &loggregator_v2.Selector_Log{Log: &loggregator_v2.LogSelector{}}},
		{Message: &loggregator_v2.Selector_Counter{Counter: &loggregator_v2.CounterSelector{}}},
		{Message: &loggregator_v2.Selector_Event{Event: &loggregator_v2.EventSelector{}}},
	}
	gaugeSelectors = []*loggregator_v2.Selector{
		{Message: &loggregator_v2.Selector_Gauge{Gauge: &loggregator_v2.GaugeSelector{}}},
		{Message: &loggregator_v2.Selector_Counter{Counter: &loggregator_v2.CounterSelector{}}},
		{Message: &loggregator_v2.Selector_Event{Event: &loggregator_v2.EventSelector{}}},
	}
	accessToken      string
	useRepRtrLogging bool
	gui              *gocui.Gui
)

func startMT(cliConnection plugin.CliConnection) {
	flaggy.DefaultParser.ShowHelpOnUnexpected = false
	flaggy.DefaultParser.ShowVersionWithVersionFlag = false
	flaggy.Bool(&useRepRtrLogging, "l", "includelogs", "Include logs from REP and RTR (more CPU overhead)")
	flaggy.Parse()
	if !conf.EnvironmentComplete(cliConnection) {
		os.Exit(8)
	}

	errorChan := make(chan error)

	go func() {
		for err := range errorChan {
			util.WriteToFile(fmt.Sprintf("from errorChannel: %s\n", err.Error()))
		}
	}()

	tokenAttacher := &TokenAttacher{}

	var err error

	go func() {
		for {
			if accessToken, err = cliConnection.AccessToken(); err != nil {
				fmt.Printf("tokenRefresher failed : %s)", err)
			}
			tokenAttacher.refreshToken(accessToken)
			time.Sleep(15 * time.Minute)
		}
	}()

	rlpGatewayClient := loggregator.NewRLPGatewayClient(
		strings.Replace(conf.ApiAddr, "api.sys", "log-stream.sys", 1),
		loggregator.WithRLPGatewayHTTPClient(tokenAttacher),
		loggregator.WithRLPGatewayErrChan(errorChan),
		loggregator.WithRLPGatewayMaxRetries(1000),
	)

	time.Sleep(1 * time.Second) // wait for uaa token to be fetched
	var envelopeStream loggregator.EnvelopeStream
	util.WriteToFile(fmt.Sprintf("useRtrRepLogging: %t", useRepRtrLogging))
	if useRepRtrLogging {
		envelopeStream = rlpGatewayClient.Stream(context.Background(), &loggregator_v2.EgressBatchRequest{ShardId: conf.ShardId, Selectors: allSelectors})
	} else {
		envelopeStream = rlpGatewayClient.Stream(context.Background(), &loggregator_v2.EgressBatchRequest{ShardId: conf.ShardId, Selectors: gaugeSelectors})
	}

	filterCache := make(map[string]bool)
	tag2filter := "origin"

	go func() {
		for {
			for _, envelope := range envelopeStream() {
				conf.TotalEnvelopes++
				orgName := envelope.Tags[apps.TagOrgName]
				spaceName := envelope.Tags[apps.TagSpaceName]
				appName := envelope.Tags[apps.TagAppName]
				index := envelope.Tags[apps.TagAppInstanceId]
				appguid := envelope.Tags[apps.TagAppId]
				key := appguid + "/" + index
				if envelopeLog := envelope.GetLog(); envelopeLog != nil {
					if envelope.Tags[apps.TagOrigin] == apps.TagOriginValueRep || envelope.Tags[apps.TagOrigin] == apps.TagOriginValueRtr {
						conf.MapLock.Lock()
						// if key not in metricMap, add it
						metricValues, ok := apps.InstanceMetricMap[key]
						if !ok {
							metricValues.Tags = make(map[string]float64)
							apps.InstanceMetricMap[key] = metricValues
						}
						if envelope.Tags[apps.TagOrigin] == apps.TagOriginValueRep {
							metricValues.LogRep++
							conf.TotalEnvelopesRep++
						}
						if envelope.Tags[apps.TagOrigin] == apps.TagOriginValueRtr {
							metricValues.LogRtr++
							conf.TotalEnvelopesRtr++
						}
						metricValues.AppName = appName
						metricValues.AppIndex = index
						metricValues.AppGuid = appguid
						metricValues.SpaceName = spaceName
						metricValues.OrgName = orgName
						metricValues.LastSeen = time.Now()
						metricValues.IP = envelope.GetTags()["ip"]
						apps.InstanceMetricMap[key] = metricValues
						conf.MapLock.Unlock()
					}
				}
				if gauge := envelope.GetGauge(); gauge != nil {
					//util.WriteToFile(fmt.Sprintf("gauge: %v / Tags: %v", gauge, envelope.GetTags()))
					if orgName != "" { // these are app-related metrics
						conf.TotalApps[appguid] = true // just count the apps (not instances)
						metrics := gauge.GetMetrics()
						indexInt, _ := strconv.Atoi(index)
						conf.MapLock.Lock()
						if indexInt+1 > apps.AppInstanceCounters[appguid].Count {
							instanceCounter := apps.AppInstanceCounter{Count: indexInt + 1, LastUpdated: time.Now()}
							apps.AppInstanceCounters[appguid] = instanceCounter
						}
						// if key not in metricMap, add it
						metricValues, ok := apps.InstanceMetricMap[key]
						if !ok {
							metricValues.Tags = make(map[string]float64)
							apps.InstanceMetricMap[key] = metricValues
						}
						for _, metricName := range apps.MetricNames {
							value := metrics[metricName].GetValue()
							if value != 0 {
								metricValues.Tags[metricName] = value
							}
						}
						metricValues.AppName = appName
						metricValues.AppIndex = index
						metricValues.AppGuid = appguid
						metricValues.SpaceName = spaceName
						metricValues.OrgName = orgName
						metricValues.LastSeen = time.Now()
						metricValues.IP = envelope.GetTags()["ip"]
						metricValues.CpuTot = metricValues.CpuTot + metricValues.Tags[apps.MetricCpu]
						apps.InstanceMetricMap[key] = metricValues
						conf.MapLock.Unlock()
					} else { // these are machine-related metrics (diego-cell / router / cc )

					}
					if envelope.Tags[apps.TagOrgName] == "" {
						tag2filter = envelope.Tags["job"] + "," + envelope.Tags["origin"] // these are cell-related metrics
						for metricKey, _ := range gauge.GetMetrics() {
							tag2filter = tag2filter + "," + metricKey
							if !filterCache[tag2filter] {
								//util.WriteToFile(fmt.Sprintf("%s", tag2filter))
								filterCache[tag2filter] = true
							}
						}
					}
				}
			}
		}
	}()

	// start up the routine that cleans up the metrics map (apps that haven't been seen for a while are removed)
	go func() {
		for range time.NewTicker(1 * time.Minute).C {
			conf.MapLock.Lock()
			var deleted = 0
			for key, metricValues := range apps.InstanceMetricMap {
				if time.Since(metricValues.LastSeen) > 1*time.Minute {
					delete(apps.InstanceMetricMap, key)
					delete(conf.TotalApps, strings.Split(key, "/")[0])           // yes we know, if multiple app instances, we will do unnecessary deletes
					delete(apps.AppInstanceCounters, strings.Split(key, "/")[0]) // yes we know, if multiple app instances, we will do unnecessary deletes
					deleted++
				}
			}
			conf.MapLock.Unlock()
		}
	}()

	// start up the routine that checks how old the value is in AppInstanceCount and lowers it if necessary
	go func() {
		for range time.NewTicker(10 * time.Second).C {
			conf.MapLock.Lock()
			for key, appInstanceCounter := range apps.AppInstanceCounters {
				if time.Since(appInstanceCounter.LastUpdated) > 30*time.Second && appInstanceCounter.Count > 1 {
					//util.WriteToFile(fmt.Sprintf("Lowered instance count for %s to %d", conf.InstanceMetricMap[key+"/0"].AppName, appInstanceCounter.Count-1))
					updatedInstanceCounter := apps.AppInstanceCounter{Count: appInstanceCounter.Count - 1, LastUpdated: time.Now()}
					apps.AppInstanceCounters[key] = updatedInstanceCounter
				}
			}
			conf.MapLock.Unlock()
		}
	}()

	StartCui()
}

// StartCui - Start the Console User Interface to present the metrics
func StartCui() {
	var err error
	gui, err = gocui.NewGui(gocui.OutputNormal, false)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer gui.Close()

	gui.SetManager(vms.NewVMView(), apps.NewAppView())

	apps.SetKeyBindings(gui)
	vms.SetKeyBindings(gui)
	common.SetKeyBindings(gui)

	//  main UI refresh loop
	go func() {
		for {
			if conf.ActiveView == conf.AppView || conf.ActiveView == conf.AppInstanceView {
				apps.ShowView(gui)
			} else {
				if conf.ActiveView == conf.VMView {
					vms.ShowView(gui)
				}
			}
		}
	}()

	if err = gui.MainLoop(); err != nil && !errors.Is(err, gocui.ErrQuit) {
		fmt.Println(err)
		gui.Close()
		os.Exit(1)
	}
}

type TokenAttacher struct {
	token string
}

func (a *TokenAttacher) refreshToken(token string) {
	a.token = token
}

func (a *TokenAttacher) Do(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", a.token)
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	return http.DefaultClient.Do(req)
}
