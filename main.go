package main

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"golang.org/x/text/language"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"golang.org/x/text/message"
	"golang.org/x/text/number"
)

const (
	formatDiffString   = "❗*Coronavirus Updates*❗\n\n*Total Cases: %v (%s%d)*\n*Total Deaths: %v (+%d)*\n*Last Updated: %s*\n\n@CoronavirusStatNews"
	formatNoDiffString = "❗*Coronavirus Updates*❗\n\n*Total Cases: %v*\n*Total Deaths: %v*\n*Last Updated: %s*"
	layout             = "Jan 2, 2006 @ 15:04 MST"
)

var (
	channelName  string
	recentRecord *Record
	bot          *tgbotapi.BotAPI

	source  *BNONewsDataSource
	printer = message.NewPrinter(language.English)
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
					sign := "+"

					if totalCasesDiff < 0 {
						sign = "-"
					}

					text = printer.Sprintf(formatDiffString, number.Decimal(newRecord.ConfirmedCases), sign, totalCasesDiff, number.Decimal(newRecord.Deaths), totalDeathsDiff, newRecord.LastUpdated.Format(layout))
				}
			} else {
				text = printer.Sprintf(formatNoDiffString, number.Decimal(newRecord.ConfirmedCases), number.Decimal(newRecord.Deaths), newRecord.LastUpdated.Format(layout))
			}

			if text != "" {
				msg := tgbotapi.MessageConfig{
					BaseChat: tgbotapi.BaseChat{
						ChannelUsername: fmt.Sprintf("@%s", channelName),
					},
					Text:      escape(text),
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

var escapes = "_[]()~`>#+-=|{}.!"

func escape(s string) string {
	var newString []rune
	for _, r := range s {
		if strings.Index(escapes, string(r)) != -1 {
			newString = append(newString, '\\')
		}
		newString = append(newString, r)
	}
	return string(newString)
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

	if f, err := os.Open("lastRecord"); err == nil {
		b, _ := ioutil.ReadAll(f)
		line := strings.Split(string(b), ",")
		cases, _ := strconv.Atoi(line[0])
		deaths, _ := strconv.Atoi(line[1])

		recentRecord = &Record{
			Province:       "",
			Country:        "Global",
			LastUpdated:    time.Now(),
			ConfirmedCases: cases,
			Deaths:         deaths,
			Recovered:      -1,
		}
	}

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

	_ = ioutil.WriteFile("lastRecord", []byte(strconv.Itoa(recentRecord.ConfirmedCases)+","+strconv.Itoa(recentRecord.Deaths)), os.ModePerm)
}
