package main

import (
	"code.cloudfoundry.org/cli/plugin"
	"code.cloudfoundry.org/go-loggregator/v10"
	"code.cloudfoundry.org/go-loggregator/v10/rpc/loggregator_v2"
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
	streamErrored    = true
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
			streamErrored = true
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

	go func() {
		for {
			if streamErrored == true { // if the stream errored (occurs quite often), we need to re-establish it
				envelopeStream = nil
				if useRepRtrLogging {
					envelopeStream = rlpGatewayClient.Stream(context.Background(), &loggregator_v2.EgressBatchRequest{ShardId: conf.ShardId, Selectors: allSelectors})
				} else {
					envelopeStream = rlpGatewayClient.Stream(context.Background(), &loggregator_v2.EgressBatchRequest{ShardId: conf.ShardId, Selectors: gaugeSelectors})
				}
				streamErrored = false
			}
			for _, envelope := range envelopeStream() {
				common.TotalEnvelopes++
				orgName := envelope.Tags[apps.TagOrgName]
				spaceName := envelope.Tags[apps.TagSpaceName]
				appName := envelope.Tags[apps.TagAppName]
				index := envelope.Tags[apps.TagAppInstanceId]
				appguid := envelope.Tags[apps.TagAppId]
				var key string
				if envelopeLog := envelope.GetLog(); envelopeLog != nil {
					key = appguid + "/" + index
					if envelope.Tags[apps.TagOrigin] == apps.TagOriginValueRep || envelope.Tags[apps.TagOrigin] == apps.TagOriginValueRtr {
						common.MapLock.Lock()
						// if key not in metricMap, add it
						metricValues, ok := apps.InstanceMetricMap[key]
						if !ok {
							metricValues.Tags = make(map[string]float64)
							apps.InstanceMetricMap[key] = metricValues
						}
						if envelope.Tags[apps.TagOrigin] == apps.TagOriginValueRep {
							metricValues.LogRep++
							common.TotalEnvelopesRep++
						}
						if envelope.Tags[apps.TagOrigin] == apps.TagOriginValueRtr {
							metricValues.LogRtr++
							common.TotalEnvelopesRtr++
						}
						metricValues.AppName = appName
						metricValues.AppIndex = index
						metricValues.AppGuid = appguid
						metricValues.SpaceName = spaceName
						metricValues.OrgName = orgName
						metricValues.LastSeen = time.Now()
						metricValues.IP = envelope.GetTags()["ip"]
						apps.InstanceMetricMap[key] = metricValues
						common.MapLock.Unlock()
					}
				}
				if gauge := envelope.GetGauge(); gauge != nil {
					//util.WriteToFile(fmt.Sprintf("gauge: %v / Tags: %v", gauge, envelope.GetTags()))
					metrics := gauge.GetMetrics()
					common.MapLock.Lock()
					if orgName != "" { // these are app-related metrics
						key = appguid + "/" + index
						apps.TotalApps[appguid] = true // just count the apps (not instances)
						indexInt, _ := strconv.Atoi(index)
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
					} else { // these are machine-related metrics (diego-cell / router / cc )
						key = envelope.Tags[vms.TagIP]
						if envelope.Tags[vms.TagIP] != "" {
							// if key not in metricMap, add it
							metricValues, ok := vms.CellMetricMap[key]
							if !ok {
								metricValues.Tags = make(map[string]float64)
								vms.CellMetricMap[key] = metricValues
							}
							for _, metricName := range vms.MetricNames {
								value := metrics[metricName].GetValue()
								if value != 0 {
									metricValues.Tags[metricName] = value
								}
							}
							metricValues.IP = envelope.Tags[vms.TagIP]
							metricValues.Job = envelope.Tags[vms.TagJob]
							metricValues.Index = envelope.Tags[vms.TagIx]
							metricValues.LastSeen = time.Now()
							vms.CellMetricMap[key] = metricValues
						}
					}
					common.MapLock.Unlock()
				}
			}
		}
	}()

	// start up the routine that cleans up the metrics map (apps that haven't been seen for a while are removed)
	go func() {
		for range time.NewTicker(1 * time.Minute).C {
			common.MapLock.Lock()
			var deleted = 0
			for key, metricValues := range apps.InstanceMetricMap {
				if time.Since(metricValues.LastSeen) > 1*time.Minute {
					delete(apps.InstanceMetricMap, key)
					delete(apps.TotalApps, strings.Split(key, "/")[0])           // yes we know, if multiple app instances, we will do unnecessary deletes
					delete(apps.AppInstanceCounters, strings.Split(key, "/")[0]) // yes we know, if multiple app instances, we will do unnecessary deletes
					deleted++
				}
			}
			common.MapLock.Unlock()
		}
	}()

	// start up the routine that checks how old the value is in AppInstanceCount and lowers it if necessary
	go func() {
		for range time.NewTicker(10 * time.Second).C {
			common.MapLock.Lock()
			for key, appInstanceCounter := range apps.AppInstanceCounters {
				if time.Since(appInstanceCounter.LastUpdated) > 30*time.Second && appInstanceCounter.Count > 1 {
					updatedInstanceCounter := apps.AppInstanceCounter{Count: appInstanceCounter.Count - 1, LastUpdated: time.Now()}
					apps.AppInstanceCounters[key] = updatedInstanceCounter
				}
			}
			common.MapLock.Unlock()
		}
	}()

	startCui()
}

// StartCui - Start the Console User Interface to present the metrics
func startCui() {
	var err error
	gui, err = gocui.NewGui(gocui.OutputNormal, false)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer gui.Close()

	//  main UI refresh loop
	go func() {
		gui.SetManager(vms.NewVMView()) // we startup with the VMView
		for streamErrored == false {
			if common.ActiveView == common.AppView || common.ActiveView == common.AppInstanceView {
				apps.SetKeyBindings(gui)
				common.SetKeyBindings(gui)
				if common.ViewToggled {
					gui.SetManager(apps.NewAppView())
					common.ViewToggled = false
				}
				apps.ShowView(gui)
			} else {
				if common.ActiveView == common.VMView {
					vms.SetKeyBindings(gui)
					common.SetKeyBindings(gui)
					if common.ViewToggled {
						gui.SetManager(vms.NewVMView())
						common.ViewToggled = false
					}
					vms.ShowView(gui)
				}
			}
			time.Sleep(time.Duration(conf.IntervalSecs) * time.Second)
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
