package tg

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/metskem/rommel/SkewnessDetector/conf"
	"log"
)

var Bot *tgbotapi.BotAPI

func SendMessage(chatId int64, message string) {
	var err error
	msgConfig := tgbotapi.MessageConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID:           chatId,
			ReplyToMessageID: 0,
		},
		Text:                  message,
		DisableWebPagePreview: true,
	}

	if Bot == nil {
		if Bot, err = tgbotapi.NewBotAPI(conf.BotToken); err != nil {
			log.Panic(err.Error())
		}
	}
	if _, err := Bot.Send(msgConfig); err != nil {
		fmt.Printf("failed to send message: %s", err)
	}

}
