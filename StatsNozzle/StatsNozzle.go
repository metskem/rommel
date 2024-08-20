package main

import (
	"context"
	"crypto/tls"
	"github.com/cloudfoundry-incubator/uaago"
	"github.com/cloudfoundry/go-cfclient/v3/config"
	"github.com/cloudfoundry/noaa/consumer"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/metskem/rommel/StatsNozzle/conf"
	"golang.org/x/oauth2"
	"log"
	"os"
	"strings"
	"time"
)

var (
	CfConfig               *config.Config
	HttpStartStopCounter   int
	CounterEventCounter    int
	LogMessageCounter      int
	ContainerMetricCounter int
	EnvelopeErrorCounter   int
	CounterEventMap        = make(map[CounterEventKey]CounterEventValue)
	ContainerMetricMap     = make(map[ContainerMetricKey]ContainerMetricValue)
)

type CounterEventKey struct {
	Name string
	//Tags  map[string]string
	Job   string
	Index string
	Ip    string
}

type CounterEventValue struct {
	Delta uint64
	Total uint64
}

func CounterEventsByName() map[string]CounterEventValue {
	countersByName := make(map[string]CounterEventValue)
	for k, v := range CounterEventMap {
		countersByName[k.Name] = v
	}
	return countersByName
}

type ContainerMetricKey struct {
	ApplicationId string
	InstanceIndex int32
}

type ContainerMetricValue struct {
	CpuPercentage    float64
	MemoryBytes      uint64
	DiskBytes        uint64
	MemoryBytesQuota uint64
	DiskBytesQuota   uint64
}

type tokenRefresher struct {
	uaaClient *uaago.Client
	user      string
	password  string
}

func (t *tokenRefresher) RefreshAuthToken() (string, error) {
	token, err := t.uaaClient.GetAuthToken(t.user, t.password, true)
	if err != nil {
		log.Fatalf("tokenRefresher failed : %s)", err)
	}
	return token, nil
}

func setCFConfig() {
	var err error
	if CfConfig, err = config.New(conf.ApiAddress, config.SkipTLSValidation(), config.ClientCredentials(conf.CfUsername, conf.CfPassword)); err != nil {
		log.Fatalf("failed to create new config: %s", err)
	}
	return
}

func printStatistics() {
	log.Printf("\nNumber of CounterEvents: %d\n", len(CounterEventMap))
	for k, v := range CounterEventsByName() {
		log.Printf("CounterEvent: %s, %d, %d\n", k, v.Delta, v.Total)
	}
	log.Printf("\nNumber of ContainerMetricEvents: %d\n", len(ContainerMetricMap))
	for k, v := range ContainerMetricMap {
		log.Printf("CounterEvent: %s(%d), %f, %d, %d\n", k.ApplicationId, k.InstanceIndex, v.CpuPercentage, v.MemoryBytes, v.DiskBytes)
	}
	//tg.SendMessage(conf.ChatId, msg)
}

func main() {
	log.SetOutput(os.Stdout)
	if !conf.EnvironmentComplete() {
		os.Exit(8)
	}

	setCFConfig()

	go func() {
		for {
			time.Sleep(time.Duration(conf.Interval) * time.Second)
			printStatistics()
		}
	}()

	go func() {
		suckFirehose(CfConfig, conf.ApiAddress, conf.CfUsername, conf.CfPassword)
	}()

	// wait forever
	time.Sleep(999999 * time.Hour)
}

func suckFirehose(cfConfig *config.Config, api, user, password string) {
	dopplerAddress := strings.Replace(api, "https://api.", "wss://doppler.", 1)
	cons := consumer.New(dopplerAddress, &tls.Config{InsecureSkipVerify: true}, nil)
	uaa, err := uaago.NewClient(strings.ReplaceAll(api, "api.sys", "uaa.sys"))
	if err != nil {
		log.Printf("error from uaaClient %s: %s\n", api, err)
		os.Exit(1)
	}
	refresher := tokenRefresher{uaaClient: uaa}
	refresher.user = user
	refresher.password = password
	cons.RefreshTokenFrom(&refresher)
	var tokenSource oauth2.TokenSource
	if tokenSource, err = cfConfig.CreateOAuth2TokenSource(context.Background()); err != nil {
		log.Fatalf("failed to create token source: %s", err)
	} else {
		var tok *oauth2.Token
		if tok, err = tokenSource.Token(); err != nil {
			log.Fatalf("failed to get token from token source: %s", err)
		} else {
			filter := consumer.Metrics
			firehoseChan, errorChan := cons.FilteredFirehose("StatsNozzle", tok.AccessToken, filter)
			go func() {
				for err = range errorChan {
					log.Printf("%v\n", err.Error())
				}
			}()
			for msg := range firehoseChan {
				switch msg.GetEventType() {
				case events.Envelope_HttpStartStop:
					HttpStartStopCounter++
				case events.Envelope_CounterEvent:
					CounterEventMap[CounterEventKey{Name: *msg.CounterEvent.Name, Job: *msg.Job, Index: *msg.Index, Ip: *msg.Ip}] = CounterEventValue{Delta: msg.CounterEvent.GetDelta(), Total: msg.CounterEvent.GetTotal()}
					CounterEventCounter++
				case events.Envelope_LogMessage:
					LogMessageCounter++
				case events.Envelope_ContainerMetric:
					ContainerMetricMap[ContainerMetricKey{ApplicationId: *msg.ContainerMetric.ApplicationId, InstanceIndex: *msg.ContainerMetric.InstanceIndex}] = ContainerMetricValue{CpuPercentage: msg.ContainerMetric.GetCpuPercentage(), MemoryBytes: msg.ContainerMetric.GetMemoryBytes(), DiskBytes: msg.ContainerMetric.GetDiskBytes(), MemoryBytesQuota: msg.ContainerMetric.GetMemoryBytesQuota(), DiskBytesQuota: msg.ContainerMetric.GetDiskBytesQuota()}
					ContainerMetricCounter++
				case events.Envelope_Error:
					EnvelopeErrorCounter++
				}
			}
		}
	}
}
