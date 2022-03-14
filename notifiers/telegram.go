package notifiers

import (
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
	"strconv"
)

var TelegramToken = os.Getenv("TELEGRAM_TOKEN")
var AdminTelegramId = os.Getenv("TARGET_USER_ID")

var BotApi *tg.BotAPI

func TestNotify(str string) {
	if BotApi == nil {
		var err error
		BotApi, err = tg.NewBotAPI(TelegramToken)
		if err != nil {
			log.Panic(err)
		}

		BotApi.Debug = false
		log.Println("Authorized on account " + BotApi.Self.UserName)
	}

	adminId, err := strconv.Atoi(AdminTelegramId)
	if err != nil {
		log.Println(err)
	}

	msg := tg.NewMessage(int64(adminId), str)
	msg.ParseMode = "HTML"
	_, err = BotApi.Send(msg)
	if err != nil {
		log.Println(err)
	}
}
