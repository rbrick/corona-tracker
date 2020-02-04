package main

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const formatString = "❗*Coronavirus Updates*❗\n\n*Total Cases: %d \\(\\+%d\\)*\n*Total Deaths: %d \\(\\+%d\\)*\n*Last Updated: %s*"

var (
	channelName  string
	recentRecord *Record
	bot          *tgbotapi.BotAPI

	source *BNONewsDataSource
)

func poll() {
	err := source.Collect()

	if err != nil {
		log.Panicln(err)
	} else {
		newRecord := source.records[0]
		if bot != nil {
			totalCasesDiff := newRecord.ConfirmedCases - recentRecord.ConfirmedCases
			totalDeathsDiff := newRecord.Deaths - recentRecord.Deaths

			if totalCasesDiff != 0 || totalDeathsDiff != 0 {
				msg := tgbotapi.MessageConfig{
					BaseChat: tgbotapi.BaseChat{
						ChannelUsername: fmt.Sprintf("@%s", channelName),
					},
					Text:      fmt.Sprintf(formatString, newRecord.ConfirmedCases, totalCasesDiff, newRecord.Deaths, totalDeathsDiff, newRecord.LastUpdated.Format("Jan 2, 2006 @ 15:04")),
					ParseMode: "markdownv2",
				}

				if _, err := bot.Send(msg); err != nil {
					log.Panicln(err)
				}
			}

		}

		recentRecord = newRecord
	}

}
func init() {
	log.Println("initializing bot")
	channelName = os.Getenv("TG_CHANNEL_NAME")
	_bot, err := tgbotapi.NewBotAPI(os.Getenv("TG_BOT_TOKEN"))
	if err != nil {
		panic(err)
	}

	bot = _bot

	log.Println("bot initialized successfully. initial polling...")
	poll()

	bno := &BNONewsDataSource{}

	_ = bno.Collect()

	log.Println("poll successful!")
}

func main() {

	ticker := time.NewTicker(5 * time.Minute)
	sigs := make(chan os.Signal)

	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)

	var wg sync.WaitGroup
	go func() {
	loop:
		for {
			select {
			case <-sigs:
				break loop
			case <-ticker.C:
				log.Println("polling...")
				poll()
			}
		}

		log.Println("Goodbye.")
		wg.Done()
	}()

	defer ticker.Stop()

	log.Println("poll loop starting. polling every 15 minutes...")
	wg.Add(1)
	wg.Wait()
}
