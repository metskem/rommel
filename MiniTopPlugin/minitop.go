package main

import (
	"code.cloudfoundry.org/cli/plugin"
	"code.cloudfoundry.org/go-loggregator/v9"
	"code.cloudfoundry.org/go-loggregator/v9/rpc/loggregator_v2"
	"context"
	"crypto/tls"
	"fmt"
	"github.com/integrii/flaggy"
	"github.com/metskem/rommel/MiniTopPlugin/conf"
	"github.com/metskem/rommel/MiniTopPlugin/cui"
	"github.com/metskem/rommel/MiniTopPlugin/util"
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
	}
	gaugeSelectors = []*loggregator_v2.Selector{
		{Message: &loggregator_v2.Selector_Gauge{Gauge: &loggregator_v2.GaugeSelector{}}},
	}
	accessToken      string
	useRepRtrLogging bool
)

func startMT(cliConnection plugin.CliConnection) {
	flaggy.DefaultParser.ShowHelpOnUnexpected = false
	flaggy.DefaultParser.ShowVersionWithVersionFlag = false
	flaggy.Bool(&useRepRtrLogging, "l", "includelogs", "Include logs from REP and RTR (excessive CPU overhead)")
	flaggy.Parse()
	if !conf.EnvironmentComplete(cliConnection) {
		os.Exit(8)
	}

	errorChan := make(chan error)

	go func() {
		for err := range errorChan {
			fmt.Printf("from errorChannel: %s\n", err.Error())
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

	c := loggregator.NewRLPGatewayClient(
		strings.Replace(conf.ApiAddr, "api.sys", "log-stream.sys", 1),
		loggregator.WithRLPGatewayHTTPClient(tokenAttacher),
		loggregator.WithRLPGatewayErrChan(errorChan),
	)

	time.Sleep(1 * time.Second) // wait for uaa token to be fetched
	var envelopeStream loggregator.EnvelopeStream
	util.WriteToFile(fmt.Sprintf("useRtrRepLogging: %t", useRepRtrLogging))
	if useRepRtrLogging {
		envelopeStream = c.Stream(context.Background(), &loggregator_v2.EgressBatchRequest{ShardId: conf.ShardId, Selectors: allSelectors})
	} else {
		envelopeStream = c.Stream(context.Background(), &loggregator_v2.EgressBatchRequest{ShardId: conf.ShardId, Selectors: gaugeSelectors})
	}

	go func() {
		for {
			for _, envelope := range envelopeStream() {
				conf.TotalEnvelopes++
				orgName := envelope.Tags[conf.TagOrgName]
				spaceName := envelope.Tags[conf.TagSpaceName]
				appName := envelope.Tags[conf.TagAppName]
				index := envelope.Tags[conf.TagAppInstanceId]
				appguid := envelope.Tags[conf.TagAppId]
				key := appguid + "/" + index
				if envelopeLog := envelope.GetLog(); envelopeLog != nil {
					if envelope.Tags[conf.TagOrigin] == conf.TagOriginValueRep || envelope.Tags[conf.TagOrigin] == conf.TagOriginValueRtr {
						conf.MapLock.Lock()
						// if key not in metricMap, add it
						metricValues, ok := conf.InstanceMetricMap[key]
						if !ok {
							metricValues.Tags = make(map[string]float64)
							conf.InstanceMetricMap[key] = metricValues
						}
						if envelope.Tags[conf.TagOrigin] == conf.TagOriginValueRep {
							metricValues.LogRep++
							conf.TotalEnvelopesRep++
						}
						if envelope.Tags[conf.TagOrigin] == conf.TagOriginValueRtr {
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
						conf.InstanceMetricMap[key] = metricValues
						conf.MapLock.Unlock()
					}
				}
				if gauge := envelope.GetGauge(); gauge != nil {
					if orgName != "" {
						conf.TotalApps[appguid] = true // just count the apps (not instances)
						metrics := gauge.GetMetrics()
						indexInt, _ := strconv.Atoi(index)
						conf.MapLock.Lock()
						if indexInt+1 > conf.AppInstanceCounters[appguid].Count {
							instanceCounter := conf.AppInstanceCounter{Count: indexInt + 1, LastUpdated: time.Now()}
							conf.AppInstanceCounters[appguid] = instanceCounter
						}
						// if key not in metricMap, add it
						metricValues, ok := conf.InstanceMetricMap[key]
						if !ok {
							metricValues.Tags = make(map[string]float64)
							conf.InstanceMetricMap[key] = metricValues
						}
						for _, metricName := range conf.MetricNames {
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
						metricValues.CpuTot = metricValues.CpuTot + metricValues.Tags[conf.MetricCpu]
						conf.InstanceMetricMap[key] = metricValues
						conf.MapLock.Unlock()
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
			for key, metricValues := range conf.InstanceMetricMap {
				if time.Since(metricValues.LastSeen) > 1*time.Minute {
					delete(conf.InstanceMetricMap, key)
					delete(conf.TotalApps, strings.Split(key, "/")[0])           // yes we know, if multiple app instances, we will do unnecessary deletes
					delete(conf.AppInstanceCounters, strings.Split(key, "/")[0]) // yes we know, if multiple app instances, we will do unnecessary deletes
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
			for key, appInstanceCounter := range conf.AppInstanceCounters {
				if time.Since(appInstanceCounter.LastUpdated) > 30*time.Second && appInstanceCounter.Count > 1 {
					//util.WriteToFile(fmt.Sprintf("Lowered instance count for %s to %d", conf.InstanceMetricMap[key+"/0"].AppName, appInstanceCounter.Count-1))
					updatedInstanceCounter := conf.AppInstanceCounter{Count: appInstanceCounter.Count - 1, LastUpdated: time.Now()}
					conf.AppInstanceCounters[key] = updatedInstanceCounter
				}
			}
			conf.MapLock.Unlock()
		}
	}()

	cui.Start()
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
