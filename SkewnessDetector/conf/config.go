package conf

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

var (
	ApiAddress       = os.Getenv("CF_API_ADDR")
	CfUsername       = os.Getenv("CF_USERNAME")
	CfPassword       = os.Getenv("CF_PASSWORD")
	CF2Password      string
	CF1Password      string
	CF2User          string
	CF1User          string
	CF2API           string
	CF1API           string
	chatIdStr        = os.Getenv("CHAT_ID")
	ChatId           int64
	BotToken         = os.Getenv("BOT_TOKEN")
	thresHoldStr     = os.Getenv("THRESHOLD")
	Threshold        float64
	thresHoldPercStr = os.Getenv("THRESHOLD_PERC")
	ThresholdPerc    float64
	intervalStr      = os.Getenv("INTERVAL")
	Interval         float64
)

func EnvironmentComplete() bool {
	envComplete := true
	if ApiAddress == "" {
		fmt.Println("missing envvar : API_ADDR")
		envComplete = false
	} else {
		CF1API = strings.Split(ApiAddress, ",")[0]
		CF2API = strings.Split(ApiAddress, ",")[1]
	}
	if CfUsername == "" {
		fmt.Println("missing envvar : CF_USERNAME")
		envComplete = false
	} else {
		CF1User = strings.Split(CfUsername, ",")[0]
		CF2User = strings.Split(CfUsername, ",")[1]
	}
	if CfPassword == "" {
		fmt.Println("missing envvar : CF_PASSWORD")
		envComplete = false
	} else {
		CF1Password = strings.Split(CfPassword, ",")[0]
		CF2Password = strings.Split(CfPassword, ",")[1]
	}
	if chatIdStr == "" {
		fmt.Println("missing envvar : CHAT_ID")
		envComplete = false
	} else {
		ChatId, _ = strconv.ParseInt(chatIdStr, 10, 64)
	}
	if BotToken == "" {
		fmt.Println("missing envvar : BOT_TOKEN")
		envComplete = false
	}
	if thresHoldStr == "" {
		fmt.Println("missing envvar : THRESHOLD")
		envComplete = false
	} else {
		Threshold, _ = strconv.ParseFloat(thresHoldStr, 64)
	}
	if thresHoldPercStr == "" {
		fmt.Println("missing envvar : THRESHOLD_PERC")
		envComplete = false
	} else {
		ThresholdPerc, _ = strconv.ParseFloat(thresHoldPercStr, 64)
	}
	if intervalStr == "" {
		Interval = 30
	} else {
		Interval, _ = strconv.ParseFloat(intervalStr, 64)
	}
	return envComplete
}
