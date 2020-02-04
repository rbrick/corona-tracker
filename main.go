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

const (
	formatDiffString   = "❗*Coronavirus Updates*❗\n\n*Total Cases: %d \\(\\+%d\\)*\n*Total Deaths: %d \\(\\+%d\\)*\n*Last Updated: %s*"
	formatNoDiffString = "❗*Coronavirus Updates*❗\n\n*Total Cases: %d*\n*Total Deaths: %d*\n*Last Updated: %s*"
	layout             = "Jan 2, 2006 @ 15:04"
)

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
			text := ""
			if recentRecord != nil {
				totalCasesDiff := newRecord.ConfirmedCases - recentRecord.ConfirmedCases
				totalDeathsDiff := newRecord.Deaths - recentRecord.Deaths

				if totalCasesDiff != 0 || totalDeathsDiff != 0 {
					text = fmt.Sprintf(formatDiffString, newRecord.ConfirmedCases, totalCasesDiff, newRecord.Deaths, totalDeathsDiff, newRecord.LastUpdated.Format(layout))
				}
			} else {
				text = fmt.Sprintf(formatNoDiffString, newRecord.ConfirmedCases, newRecord.Deaths, newRecord.LastUpdated.Format(layout))
			}

			msg := tgbotapi.MessageConfig{
				BaseChat: tgbotapi.BaseChat{
					ChannelUsername: fmt.Sprintf("@%s", channelName),
				},
				Text:      text,
				ParseMode: "markdownv2",
			}

			if _, err := bot.Send(msg); err != nil {
				log.Panicln(err)
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

	source = &BNONewsDataSource{}

	log.Println("bot initialized successfully. initial polling...")
	poll()
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
