package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ejuravlev/statham_bot/db"
	settings "github.com/ejuravlev/statham_bot/settings"
	"github.com/go-co-op/gocron/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func initPollingBotApi(settings *settings.AppSettings) *tgbotapi.BotAPI {
	api, err := tgbotapi.NewBotAPI(settings.Bot.Token)
	if err != nil {
		log.Panicln(err)
		os.Exit(1)
	}

	api.Debug = settings.Debug
	return api
}

func initWebhookBotApi(settings *settings.AppSettings) *tgbotapi.BotAPI {
	api, err := tgbotapi.NewBotAPI(settings.Bot.Token)
	if err != nil {
		log.Fatal(err)
	}

	api.Debug = settings.Debug

	wh, _ := tgbotapi.NewWebhook(fmt.Sprintf("%s/%s", settings.Bot.WebhookBaseUrl, settings.Bot.Token))

	_, err = api.Request(wh)
	if err != nil {
		log.Fatalln(err)
	}

	info, err := api.GetWebhookInfo()
	if err != nil {
		log.Fatal(err)
	}

	if info.LastErrorDate != 0 {
		log.Printf("Telegram callback failed: %s", info.LastErrorMessage)
	}
	return api
}

func getQuote(database *db.Db) string {
	quote := database.GetRandomStathamQuote()
	return fmt.Sprintf("Цитата дня #%d\n%s", quote.Id, quote.Text)
}

func setSubscription(database *db.Db, chatId int64, subscribed bool) error {
	if subscribed {
		return database.SubscribeChat(chatId)
	}
	return database.UnsubscribeChat(chatId)
}

func main() {
	settings := settings.GetSettings()

	var bot *tgbotapi.BotAPI
	var updates tgbotapi.UpdatesChannel

	log.Default().Println("Application is ready")

	database, err := db.NewConnection(settings.Db.ConnectionString)
	if err != nil {
		log.Fatalln("Can't connect to database", err)
		os.Exit(1)
	}

	if settings.Bot.UsePolling {
		bot = initPollingBotApi(&settings)
		u := tgbotapi.NewUpdate(0)
		u.Timeout = 60
		updates = bot.GetUpdatesChan(u)
	} else {
		bot = initWebhookBotApi(&settings)
		updates = bot.ListenForWebhook("/" + bot.Token)
		go http.ListenAndServe("0.0.0.0:8111", nil)
	}

	go func() {
		for update := range updates {
			if update.Message == nil {
				continue
			}

			if !update.Message.IsCommand() {
				continue
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

			processingError := error(nil)

			switch update.Message.Command() {
			case "help":
				msg.Text = "Клацни /quote чтобы получить мудрость"
			case "quote":
				msg.Text = getQuote(database)
			case "subscribe":
				processingError = setSubscription(database, update.Message.Chat.ID, true)
				msg.Text = "Подписка на цитаты оформлена"
			case "unsubscribe":
				processingError = setSubscription(database, update.Message.Chat.ID, false)
				msg.Text = "Подписка на циататы отменена"
			default:
				msg.Text = "Не понял тебя, брат"
			}

			if processingError != nil {
				log.Default().Printf("can't process user %d request: %e\n", update.Message.Chat.ID, processingError)
			}

			if _, err := bot.Send(msg); err != nil {
				log.Panic(err)
			}
		}
	}()

	// subs, err := database.GetSubscibers()
	// if err != nil {
	// 	log.Default().Fatalf("error while doing cron job: %e", err)
	// 	return
	// }
	// fmt.Println(subs)

	jobDefinition := gocron.CronJob("0 9 * * *", false)
	if settings.Debug {
		jobDefinition = gocron.DurationJob(
			10 * time.Second,
		)
	}

	scheduler, _ := gocron.NewScheduler()
	job, err := scheduler.NewJob(
		jobDefinition,
		gocron.NewTask(
			func() {
				subs, err := database.GetSubscibers()
				if err != nil {
					log.Default().Fatalf("error while doing cron job: %e", err)
					return
				}

				for _, id := range subs {
					bot.Send(tgbotapi.NewMessage(id, getQuote(database)))
				}
			},
		),
	)

	if err != nil {
		log.Fatalln(err)
	}

	log.Println(job.ID())

	scheduler.Start()

	<-time.After(9223372036854775807 * time.Nanosecond)

	err = scheduler.Shutdown()
	if err != nil {
		os.Exit(1)
	}

}
