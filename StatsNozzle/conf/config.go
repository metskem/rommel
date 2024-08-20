package conf

import (
	"fmt"
	"os"
	"strconv"
)

var (
	ApiAddress  = os.Getenv("CF_API_ADDR")
	CfUsername  = os.Getenv("CF_USERNAME")
	CfPassword  = os.Getenv("CF_PASSWORD")
	chatIdStr   = os.Getenv("CHAT_ID")
	ChatId      int64
	BotToken    = os.Getenv("BOT_TOKEN")
	intervalStr = os.Getenv("INTERVAL")
	Interval    float64
)

func EnvironmentComplete() bool {
	envComplete := true
	if ApiAddress == "" {
		fmt.Println("missing envvar : API_ADDR")
		envComplete = false
	}
	if CfUsername == "" {
		fmt.Println("missing envvar : CF_USERNAME")
		envComplete = false
	}
	if CfPassword == "" {
		fmt.Println("missing envvar : CF_PASSWORD")
		envComplete = false
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
	if intervalStr == "" {
		Interval = 30
	} else {
		Interval, _ = strconv.ParseFloat(intervalStr, 64)
	}
	return envComplete
}
