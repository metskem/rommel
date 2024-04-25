package main

import (
	"crypto/tls"
	"fmt"
	"github.com/cloudfoundry-community/go-cfclient/v3/config"
	"github.com/cloudfoundry-incubator/uaago"
	"github.com/cloudfoundry/noaa/consumer"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/metskem/rommel/SkewnessDetector/conf"
	"github.com/metskem/rommel/SkewnessDetector/tg"
	"log"
	"math"
	"os"
	"strings"
	"time"
)

var (
	CF2CfConfig *config.Config
	CF1CfConfig *config.Config
	CF2Count    int
	CF1Count    int
)

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
	if CF1CfConfig, err = config.NewClientSecret(conf.CF1API, conf.CF1User, conf.CF1Password); err != nil {
		log.Fatalf("failed to create new config: %s", err)
	} else {
		CF1CfConfig.WithSkipTLSValidation(true)
	}
	if CF2CfConfig, err = config.NewClientSecret(conf.CF2API, conf.CF2User, conf.CF2Password); err != nil {
		log.Fatalf("failed to create new config: %s", err)
	} else {
		CF2CfConfig.WithSkipTLSValidation(true)
	}
	return
}

func printStatistics() {
	var diff, total float64
	diff = math.Abs(float64(CF1Count - CF2Count))
	total = float64(CF1Count + CF2Count)
	percentage := diff / total * 100
	msg := fmt.Sprintf("CF1: %d, CF2: %d (diff: %.f, total: %.f, perc: %.f)\n", CF1Count, CF2Count, diff, total, percentage)
	log.Print(msg)
	if percentage > conf.ThresholdPerc && total/conf.Interval > conf.Threshold {
		tg.SendMessage(conf.ChatId, msg)
	}
	CF1Count = 0
	CF2Count = 0
}

func main() {
	if !conf.EnvironmentComplete() {
		os.Exit(8)
	}

	setCFConfig()
	//
	// start sucking the firehoses and handle events
	//

	go func() {
		for {
			time.Sleep(time.Duration(conf.Interval) * time.Second)
			printStatistics()
		}
	}()

	go func() {
		suckFirehose(CF1CfConfig, conf.CF1API, conf.CF1User, conf.CF1Password, &CF1Count)
	}()

	go func() {
		suckFirehose(CF2CfConfig, conf.CF2API, conf.CF2User, conf.CF2Password, &CF2Count)
	}()

	// wait forever
	time.Sleep(999999 * time.Hour)
}

func suckFirehose(cfConfig *config.Config, api, user, password string, counter *int) {
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
	firehoseChan, errorChan := cons.Firehose("StatsNozzle", cfConfig.AccessToken)

	go func() {
		for err := range errorChan {
			log.Printf("%v\n", err.Error())
		}
	}()

	for msg := range firehoseChan {
		if msg.GetEventType() == events.Envelope_HttpStartStop {
			*counter++
		}
	}
}
