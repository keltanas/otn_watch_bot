package main

import (
	"os"
	"log"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func main()  {
	apiToken := os.Getenv("API_TOKEN")
	if "" == apiToken {
		log.Println("You should expect API_TOKEN env variable")
		message, err := getData()
		log.Println(message)
		if nil != err {
			log.Println(err)
		}
		return
	}

	debug := false
	debugEnv := os.Getenv("DEBUG")
	if "true" == debugEnv || "1" == debugEnv {
		debug = true
	}

	bot, err := tgbotapi.NewBotAPI(apiToken)
	if nil != err {
		log.Panic("Cannot create bot", err)
	}

	bot.Debug = debug
	var u tgbotapi.UpdateConfig = tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	for {
		select {
		case update := <- updates:
			if nil == update.Message {
				continue
			}

			switch update.Message.Text {
			case "/rate":
				message, err := getData()

				if nil != err {
					log.Print(err)
					continue
				}

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, message)
				msg.ReplyToMessageID = update.Message.MessageID

				bot.Send(msg)
			}
		}
	}
}
