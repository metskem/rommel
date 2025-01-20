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
	"github.com/metskem/rommel/MiniTopPlugin/routes"
	"github.com/metskem/rommel/MiniTopPlugin/util"
	"github.com/metskem/rommel/MiniTopPlugin/vms"
	"net/http"
	"net/url"
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
		//{Message: &loggregator_v2.Selector_Timer{Timer: &loggregator_v2.TimerSelector{}}}, // timer events are only http request timings
		//{Message: &loggregator_v2.Selector_Event{Event: &loggregator_v2.EventSelector{}}}, // produces nothing
	}
	gaugeSelectors = []*loggregator_v2.Selector{
		{Message: &loggregator_v2.Selector_Gauge{Gauge: &loggregator_v2.GaugeSelector{}}},
		//{Message: &loggregator_v2.Selector_Log{Log: &loggregator_v2.LogSelector{}}},
		{Message: &loggregator_v2.Selector_Counter{Counter: &loggregator_v2.CounterSelector{}}},
		{Message: &loggregator_v2.Selector_Timer{Timer: &loggregator_v2.TimerSelector{}}}, // timer events are only http request timings
		//{Message: &loggregator_v2.Selector_Event{Event: &loggregator_v2.EventSelector{}}}, // produces nothing
	}
	useRepRtrLogging bool
	gui              *gocui.Gui
)

func startMT(cliConnection plugin.CliConnection) {
	flaggy.DefaultParser.ShowHelpOnUnexpected = false
	flaggy.DefaultParser.ShowVersionWithVersionFlag = false
	flaggy.Bool(&useRepRtrLogging, "l", "includelogs", "Include logs from REP and RTR (more CPU overhead)")
	flaggy.Bool(&conf.UseDebugging, "d", "debug", "Turn debugging on/off")
	flaggy.Parse()
	if !conf.EnvironmentComplete(cliConnection) {
		os.Exit(8)
	}

	errorChan := make(chan error)

	rlpCtx := context.TODO()

	tokenAttacher := NewTokenAttacher(cliConnection)

	go func() {
		for err := range errorChan {
			util.WriteToFile(fmt.Sprintf("from errorChannel: %s\n", err.Error()))
			tokenAttacher.refreshToken(cliConnection) // the most common reason for errors is that the token has expired
		}
	}()

	time.Sleep(1 * time.Second) // wait for uaa token to be fetched

	rlpGatewayClient := loggregator.NewRLPGatewayClient(
		strings.Replace(conf.ApiAddr, "api.sys", "log-stream.sys", 1),
		loggregator.WithRLPGatewayHTTPClient(tokenAttacher),
		loggregator.WithRLPGatewayErrChan(errorChan),
	)

	var envelopeStream loggregator.EnvelopeStream
	if useRepRtrLogging {
		envelopeStream = rlpGatewayClient.Stream(rlpCtx, &loggregator_v2.EgressBatchRequest{ShardId: conf.ShardId, Selectors: allSelectors})
	} else {
		envelopeStream = rlpGatewayClient.Stream(rlpCtx, &loggregator_v2.EgressBatchRequest{ShardId: conf.ShardId, Selectors: gaugeSelectors})
	}

	go func() {
		for {
			for _, envelope := range envelopeStream() {
				common.MapLock.Lock()
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
					}
				}
				if gauge := envelope.GetGauge(); gauge != nil {
					metrics := gauge.GetMetrics()
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
					} else {
						// these are machine-related metrics (diego-cell / router / cc )
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
							metricValues.LastSeen = time.Now()
							vms.CellMetricMap[key] = metricValues
						}
					}
				}
				// type counter metrics
				if counter := envelope.GetCounter(); counter != nil {
					//if _, ok := countersMap[counter.Name]; !ok {
					//	countersMap[counter.Name] = counter.Name
					//	util.WriteToFile(fmt.Sprintf("%s: %d/%d [%s, %s, %s, %s]", counter.Name, counter.Delta, counter.Total, envelope.Tags[vms.TagOrigin], envelope.Tags[vms.TagJob], envelope.Tags[vms.TagIP], envelope.Tags[apps.TagAppName]))
					//}
					key = envelope.Tags[vms.TagIP]
					if envelope.Tags[vms.TagIP] != "" {
						// if key not in metricMap, add it
						metricValues, ok := vms.CellMetricMap[key]
						if !ok {
							metricValues.Tags = make(map[string]float64)
							vms.CellMetricMap[key] = metricValues
						}
						for _, metricName := range vms.MetricNames {
							if counter.Name == metricName {
								metricValues.Tags[metricName] = metricValues.Tags[metricName] + float64(counter.Delta)
								if metricValues.Tags[metricName] == 0 {
									// it might be that this counter only has Total, not Delta
									metricValues.Tags[metricName] = float64(counter.Total)
								}
							}
						}
						metricValues.IP = envelope.Tags[vms.TagIP]
						metricValues.Job = envelope.Tags[vms.TagJob]
						metricValues.LastSeen = time.Now()
						vms.CellMetricMap[key] = metricValues
					}
				}
				// type counter metrics
				//
				//  a new URI view, showing uri (or just hostname), with req count, by status code
				//  a new remoteClient view, showing remote client IP, with req count, by status code
				if timer := envelope.GetTimer(); timer != nil && timer.Name == "http" {
					if envelope.Tags[routes.TagUri] != "" {
						if Url, err := url.Parse(envelope.Tags[routes.TagUri]); err == nil {
							key = Url.Host
							routeMetric, ok := routes.RouteMetricMap[key]
							if !ok {
								routeMetric = routes.RouteMetric{Host: key}
							}
							routeMetric.LastSeen = time.Now()
							if strings.HasPrefix(envelope.Tags[routes.TagStatusCode], "2") {
								routeMetric.R2xx++
							}
							if strings.HasPrefix(envelope.Tags[routes.TagStatusCode], "3") {
								routeMetric.R3xx++
							}
							if strings.HasPrefix(envelope.Tags[routes.TagStatusCode], "4") {
								routeMetric.R4xx++
							}
							if strings.HasPrefix(envelope.Tags[routes.TagStatusCode], "5") {
								routeMetric.R5xx++
							}
							if envelope.Tags[routes.TagMethod] == "GET" {
								routeMetric.GETs++
							}
							if envelope.Tags[routes.TagMethod] == "PUT" {
								routeMetric.PUTs++
							}
							if envelope.Tags[routes.TagMethod] == "POST" {
								routeMetric.POSTs++
							}
							if envelope.Tags[routes.TagMethod] == "DELETE" {
								routeMetric.DELETEs++
							}
							routeMetric.RTotal++
							routeMetric.TotalRespTime = routeMetric.TotalRespTime + timer.Stop - timer.Start
							routes.RouteMetricMap[key] = routeMetric
							//util.WriteToFile(fmt.Sprintf("%s: %s %d %d %d %d %d", envelope.Tags["job"], key, len(routes.RouteMetricMap), routes.RouteMetricMap[key].R2xx, routes.RouteMetricMap[key].R4xx, routes.RouteMetricMap[key].POSTs, routes.RouteMetricMap[key].TotalRespTime))
							routes.RouteMetricMap[key] = routeMetric
						}
					}
				}
				common.MapLock.Unlock()
			}
		}
	}()

	// start up the routine that cleans up the metrics map (apps that haven't been seen for a while are removed)
	go func() {
		util.WriteToFileDebug("starting app metric cleanup")
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
		util.WriteToFileDebug("starting instance metric cleanup")
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
	util.WriteToFileDebug("starting CUI")
	var err error
	gui, err = gocui.NewGui(gocui.OutputNormal, false)
	if err != nil {
		util.WriteToFile(fmt.Sprintf("failed to start CUI: %s", err))
		os.Exit(1)
	}
	defer gui.Close()

	//  main UI refresh loop
	go func() {
		util.WriteToFileDebug("starting main UI refresh loop")
		gui.SetManager(vms.NewVMView()) // we startup with the VMView
		for {
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
		util.WriteToFile(fmt.Sprintf("error in mainLoop: %s", err))
		gui.Close()
		os.Exit(1)
	}
}

func NewTokenAttacher(cliConnection plugin.CliConnection) *TokenAttacher {
	ta := &TokenAttacher{cliConnection: cliConnection}
	ta.refreshToken(cliConnection)
	return ta
}

type TokenAttacher struct {
	token         string
	calls         int
	cliConnection plugin.CliConnection
}

func (ta *TokenAttacher) refreshToken(cliConnection plugin.CliConnection) {
	if oauthToken, err := cliConnection.CliCommandWithoutTerminalOutput("oauth-token"); err != nil {
		util.WriteToFile(fmt.Sprintf("oauth-token failed : %s)", err))
	} else {
		token := strings.Fields(oauthToken[0])[1]
		ta.token = token
		util.WriteToFileDebug(fmt.Sprintf("oauth token refreshed: %s", token[len(token)-10:]))
	}
}

// Do - attach the token to the request, called once a minute
func (ta *TokenAttacher) Do(req *http.Request) (*http.Response, error) {
	ta.calls++
	if !util.IsTokenValid(ta.token) {
		ta.refreshToken(ta.cliConnection)
	}
	util.WriteToFileDebug(fmt.Sprintf("TokenAttacher.Do called %d times, token: %s", ta.calls, ta.token[len(ta.token)-10:]))
	req.Header.Set("Authorization", ta.token)
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	return http.DefaultClient.Do(req)
}
