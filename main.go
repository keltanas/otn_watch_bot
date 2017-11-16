package main

import (
	"os"
	"log"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"net/http"
	"io/ioutil"
	"encoding/json"
)

func main()  {
	apiToken := os.Getenv("API_TOKEN")
	if "" == apiToken {
		log.Print("You should expect API_TOKEN env variable")
		return
	}

	bot, err := tgbotapi.NewBotAPI(apiToken)
	if nil != err {
		log.Panic("Cannot create bot", err)
	}

	bot.Debug = true
	var u tgbotapi.UpdateConfig = tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	for {
		select {
		case update := <- updates:
			if nil == update.Message {
				continue
			}
			//log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			if "/rate" != update.Message.Text {
				continue
			}

			response, err := http.Get("https://api.livecoin.net/exchange/maxbid_minask")
			if nil != err {
				log.Print(err)
				continue
			}

			data, err := ioutil.ReadAll(response.Body)
			var result currencyPairs
			var message string

			{
				err := json.Unmarshal(data, &result)
				if nil != err {
					log.Print(err)
					continue
				}
			}

			for _, v := range result.CurrencyPairs {
				if v.Symbol == "OTN/USD" || v.Symbol == "OTN/BTC" || v.Symbol == "OTN/ETH" {
					message += "Symbol: " + v.Symbol + "\n"
					message += "Bid: " + v.MaxBid + "\n"
					message += "Ask: " + v.MinAsk + "\n"
					message += "\n"
				}
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, message)
			msg.ReplyToMessageID = update.Message.MessageID

			bot.Send(msg)
		}
	}
}
