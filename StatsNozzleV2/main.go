package main

import (
	"code.cloudfoundry.org/go-loggregator/v9"
	"code.cloudfoundry.org/go-loggregator/v9/rpc/loggregator_v2"
	"context"
	"crypto/tls"
	"github.com/cloudfoundry-incubator/uaago"
	"github.com/metskem/rommel/StatsNozzleV2/conf"
	"github.com/metskem/rommel/StatsNozzleV2/cui"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strings"
	"time"
)

var (
	allSelectors = []*loggregator_v2.Selector{
		{Message: &loggregator_v2.Selector_Gauge{Gauge: &loggregator_v2.GaugeSelector{}}},
		{Message: &loggregator_v2.Selector_Log{Log: &loggregator_v2.LogSelector{}}},
	}
	accessToken string
)

func main() {
	if !conf.EnvironmentComplete() {
		os.Exit(8)
	}

	errorChan := make(chan error)

	go func() {
		for err := range errorChan {
			log.Printf("from errorChannel: %s\n", err.Error())
		}
	}()

	uaa, err := uaago.NewClient(strings.Replace(conf.ApiAddr, "api.sys", "uaa.sys", 1))
	if err != nil {
		log.Printf("error while getting uaaClient %s\n", err)
		os.Exit(1)
	}

	tokenAttacher := &TokenAttacher{}

	go func() {
		for {
			if accessToken, err = uaa.GetAuthToken(conf.Client, conf.Secret, true); err != nil {
				log.Fatalf("tokenRefresher failed : %s)", err)
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
	envelopeStream := c.Stream(context.Background(), &loggregator_v2.EgressBatchRequest{ShardId: conf.ShardId, Selectors: allSelectors})

	go func() {
		for {
			for _, envelope := range envelopeStream() {
				conf.TotalEnvelopes++
				orgName := envelope.Tags[conf.TagOrgName]
				spaceName := envelope.Tags[conf.TagSpaceName]
				appName := envelope.Tags[conf.TagAppName]
				index := envelope.InstanceId
				appguid := envelope.SourceId
				key := appguid + "/" + index
				if envelopeLog := envelope.GetLog(); envelopeLog != nil {
					if envelope.Tags[conf.TagOrigin] == conf.TagOriginValueRep || envelope.Tags[conf.TagOrigin] == conf.TagOriginValueRtr {
						// if key not in metricMap, add it
						conf.MapLock.Lock()
						metricValues, ok := conf.MetricMap[key]
						if !ok {
							metricValues.Values = make(map[string]float64)
							conf.MetricMap[key] = metricValues
						}
						if envelope.Tags[conf.TagOrigin] == conf.TagOriginValueRep {
							metricValues.LogRep++
						}
						if envelope.Tags[conf.TagOrigin] == conf.TagOriginValueRtr {
							metricValues.LogRtr++
						}
						metricValues.AppName = appName
						metricValues.AppIndex = index
						metricValues.SpaceName = spaceName
						metricValues.OrgName = orgName
						conf.MetricMap[key] = metricValues
						conf.MapLock.Unlock()
					}
				}
				if gauge := envelope.GetGauge(); gauge != nil {
					if orgName != "" {
						metrics := gauge.GetMetrics()
						// if key not in metricMap, add it
						conf.MapLock.Lock()
						metricValues, ok := conf.MetricMap[key]
						if !ok {
							metricValues.Values = make(map[string]float64)
							conf.MetricMap[key] = metricValues
						}
						for _, metricName := range conf.MetricNames {
							value := metrics[metricName].GetValue()
							if value != 0 {
								metricValues.Values[metricName] = value
							}
						}
						metricValues.AppName = appName
						metricValues.AppIndex = index
						metricValues.SpaceName = spaceName
						metricValues.OrgName = orgName
						metricValues.CpuTot = metricValues.CpuTot + metricValues.Values[conf.MetricCpu]
						conf.MetricMap[key] = metricValues
						conf.MapLock.Unlock()
					}
				}
			}
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
