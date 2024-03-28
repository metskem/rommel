package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/go-auxiliaries/selenium"
	"github.com/go-auxiliaries/selenium/chrome"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/metskem/rommel/joeradiostats/conf"
	"github.com/metskem/rommel/joeradiostats/db"
	"github.com/metskem/rommel/joeradiostats/model"
	"github.com/metskem/rommel/joeradiostats/util"
	"log"
	"os"
	"strings"
	"time"
)

var joeURL = "https://joe.nl"

func main() {

	conf.EnvironmentComplete()
	log.SetOutput(os.Stdout)
	db.Initdb()

	if service, err := selenium.NewChromeDriverService(conf.ChromeDriverPath, 4444); err != nil {
		log.Fatal("Error:", err)
	} else {
		defer func() { _ = service.Stop() }()
		prefs := make(map[string]interface{})
		prefs["profile.managed_default_content_settings.images"] = 2
		caps := selenium.Capabilities{}
		caps.AddChrome(chrome.Capabilities{Args: []string{"--headless"}, Prefs: prefs})

		var driver selenium.WebDriver
		if driver, err = selenium.NewRemote(caps, ""); err != nil {
			log.Fatal("Error:", err)
		}
		if err = driver.Get(joeURL); err != nil {
			log.Fatal("Error:", err)
		}

		time.Sleep(2 * time.Second)

		var element selenium.WebElement
		if element, err = driver.FindElement(selenium.ByID, "pg-shadow-host"); err != nil {
			log.Fatalf("could not find element with id pg-shadow-root : %s", err)
		}
		var shadowRoot selenium.ShadowRoot
		if shadowRoot, err = element.GetElementShadowRoot(); err != nil {
			log.Fatalf("could not find html template: %s", err)
		}
		if element, err = shadowRoot.FindElement(selenium.ByID, "pg-accept-btn"); err != nil {
			log.Fatalf("could not find element with id pg-accept-bin: %s", err)
		}
		if err = element.Click(); err != nil {
			log.Fatalf("click button failed: %s", err)
		}

		time.Sleep(2 * time.Second)

		//
		// main loop routine
		go func() {
			var existingSong model.Song
			for {
				if err = driver.Get(joeURL); err != nil {
					log.Fatalf("Failed to get joeURL %s: %s", joeURL, err)
				}
				time.Sleep(2 * time.Second)
				artist, song := getSong(driver)
				log.Printf("%s  |  %s\n", artist, song)
				if artist != "" && song != "" {
					if existingSong, err = db.GetSong(artist, song); err != nil && !errors.Is(err, sql.ErrNoRows) {
						log.Printf("failed to get song %s by %s from db: %s", song, artist, err)
					} else {
						if existingSong.Id == 0 {
							newId := db.InsertSong(model.Song{Artist: artist, Title: song})
							db.InsertPlayMoment(newId)
						} else {
							if existingSong.LastPlayed.Add(1 * time.Hour).Before(time.Now()) {
								db.InsertPlayMoment(existingSong.Id)
							}
						}
					}
				}
				time.Sleep(3 * time.Minute)
			}
		}()

		//
		handleTelegramChannel()
		//
	}
}

func handleTelegramChannel() {
	var err error
	if util.Bot, err = tgbotapi.NewBotAPI(conf.BotToken); err != nil {
		log.Panic(err.Error())
	}

	util.Bot.Debug = conf.Debug

	meDetails := fmt.Sprintf("BOT: ID:%d UserName:%s FirstName:%s LastName:%s", util.Bot.Self.ID, util.Bot.Self.UserName, util.Bot.Self.FirstName, util.Bot.Self.LastName)
	log.Printf("Started Bot: %s, version:%s, build time:%s, commit hash:%s", meDetails, conf.VersionTag, conf.BuildTime, conf.CommitHash)

	newUpdate := tgbotapi.NewUpdate(0)
	newUpdate.Timeout = 60

	updatesChan, err := util.Bot.GetUpdatesChan(newUpdate)
	if err == nil {

		// start listening for messages, and optionally respond
		for update := range updatesChan {
			if update.Message == nil { // ignore any non-Message Updates
				log.Println("ignored null update")
			} else {
				chat := update.Message.Chat
				mentionedMe, cmdMe := util.TalkOrCmdToMe(update)

				// check if someone is talking to me:
				if (chat.IsPrivate() || (chat.IsGroup() && mentionedMe)) && update.Message.Text != "/start" {
					log.Printf("[%s] [chat:%d] %s\n", update.Message.From.UserName, chat.ID, update.Message.Text)
					//if cmdMe {
					//	fromUser := update.Message.From.UserName
					//	if chat.IsPrivate() {
					//		fromUser = chat.UserName
					//	}
					//	if _, err := util.Bot.Send(tgbotapi.MessageConfig{BaseChat: tgbotapi.BaseChat{ChatID: chat.ID, ReplyToMessageID: 0}, Text: fmt.Sprintf("Welcome %s!", fromUser), DisableWebPagePreview: true}); err != nil {
					//		log.Printf("failed sending message to chat %d, error is %v", chat.ID, err)
					//	}
					//}
				}

				// check if someone started a new chat
				if chat.IsPrivate() && cmdMe && update.Message.Text == "/start" {
					log.Printf("new chat added, chatid: %d, chat: %s (%s %s)\n", chat.ID, chat.UserName, chat.FirstName, chat.LastName)
				}

				// Top artists most songs
				if chat.IsPrivate() && cmdMe && update.Message.Text == "/topartistsmostsongs" {
					log.Printf("topartistsmostsongs requested, chatid: %d, chat: %s (%s %s)\n", chat.ID, chat.UserName, chat.FirstName, chat.LastName)
					if top, err := db.GetTopArtistsMostSongs(); err != nil {
						log.Printf("failed getting topartistsmostsongs: %v", err)
					} else {
						msg := "Top 10 artists with most songs:\n"
						for _, row := range top {
							msg += fmt.Sprintf("%d: %s\n", row.Count, row.Artist)
						}
						if _, err = util.Bot.Send(tgbotapi.MessageConfig{BaseChat: tgbotapi.BaseChat{ChatID: chat.ID, ReplyToMessageID: 0}, Text: msg, DisableWebPagePreview: true}); err != nil {
							log.Printf("failed sending message to chat %d, error is %v", chat.ID, err)
						}
					}
				}

				// Top 10 artists most often played
				if chat.IsPrivate() && cmdMe && update.Message.Text == "/topartistsmostoftenplayed" {
					log.Printf("topartistsmostoftenplayed requested, chatid: %d, chat: %s (%s %s)\n", chat.ID, chat.UserName, chat.FirstName, chat.LastName)
					if top, err := db.GetTopArtistsMostOftenPlayed(); err != nil {
						log.Printf("failed getting topartistsmostoftenplayed: %v", err)
					} else {
						msg := "Top 10 artists most often played:\n"
						for _, row := range top {
							msg += fmt.Sprintf("%d: %s\n", row.Count, row.Artist)
						}
						if _, err = util.Bot.Send(tgbotapi.MessageConfig{BaseChat: tgbotapi.BaseChat{ChatID: chat.ID, ReplyToMessageID: 0}, Text: msg, DisableWebPagePreview: true}); err != nil {
							log.Printf("failed sending message to chat %d, error is %v", chat.ID, err)
						}
					}
				}

				// Top duplicates
				if chat.IsPrivate() && cmdMe && update.Message.Text == "/topduplicates" {
					log.Printf("totals requested, chatid: %d, chat: %s (%s %s)\n", chat.ID, chat.UserName, chat.FirstName, chat.LastName)
					if topDup, err := db.GetTopDuplicates(); err != nil {
						log.Printf("failed getting topduplicates: %v", err)
					} else {
						msg := "Top 20 duplicates:\n"
						for _, row := range topDup {
							msg += fmt.Sprintf("%d: %s | %s\n", row.Count, row.Artist, row.Title)
						}
						if _, err = util.Bot.Send(tgbotapi.MessageConfig{BaseChat: tgbotapi.BaseChat{ChatID: chat.ID, ReplyToMessageID: 0}, Text: msg, DisableWebPagePreview: true}); err != nil {
							log.Printf("failed sending message to chat %d, error is %v", chat.ID, err)
						}
					}
				}

				// Totals
				if chat.IsPrivate() && cmdMe && update.Message.Text == "/totals" {
					log.Printf("totals requested, chatid: %d, chat: %s (%s %s)\n", chat.ID, chat.UserName, chat.FirstName, chat.LastName)
					if cnt1, cnt2, cnt3, cnt4, err := db.GetTotals(); err != nil {
						log.Printf("failed getting totals: %v", err)
					} else {
						msg := fmt.Sprintf("Total songs played: %d\nTotal unique songs: %d\nTotal unique artists: %d\nSongs played once: %d\n", cnt1, cnt2, cnt3, cnt4)
						if _, err = util.Bot.Send(tgbotapi.MessageConfig{BaseChat: tgbotapi.BaseChat{ChatID: chat.ID, ReplyToMessageID: 0}, Text: msg, DisableWebPagePreview: true}); err != nil {
							log.Printf("failed sending message to chat %d, error is %v", chat.ID, err)
						}
					}
				}
			}
		}
		fmt.Println("")

	} else {
		log.Printf("failed getting Bot updatesChannel, error: %v", err)
		os.Exit(8)
	}
}

// getSong returns the artist and song name from the given webDriver (html parsing it). Returns 2 blanks strings if not found.
func getSong(driver selenium.WebDriver) (artist string, song string) {
	elementType := "div"
	var webElements []selenium.WebElement
	var err error
	var text string
	if webElements, err = driver.FindElements(selenium.ByCSSSelector, elementType); err != nil {
		log.Printf("could not find Element type %s: %s", elementType, err)
	}
	for ix, element := range webElements {
		if text, err = element.Text(); err == nil {
			if ix == 8 {
				tokens := strings.Split(text, "\n")
				if len(tokens) >= 2 {
					song = tokens[0]
					artist = tokens[1]
				}
			}
		}
	}
	return artist, song
}
