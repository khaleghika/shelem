package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

var db map[int64]*Game

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file:", err)
	}

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("bot token not found in env")
	}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	// پردازش پیام‌ها
	for update := range updates {

		if update.Message == nil { // skip any non-Message Updates
			continue
		}

		if update.Message.Chat == nil {
			continue
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "به بات خوش آمدید!")
		bot.Send(msg)

		state, ok := db[update.Message.Chat.ID]

		if !ok {
			state = &Game{}
			db[update.Message.Chat.ID] = state
			if update.Message.Text == "/start" {
				m := createStartGame(update.Message.Chat.ID, update.Message.MessageID)
				bot.Send(m)
			} else {
				m := createStartGameWithError(update.Message.Chat.ID, update.Message.MessageID)
				bot.Send(m)
			}
		}

		state.State = processUpdate(state, update)
		msg := sendResponse(state.State, update.Message.Chat.ID, update.Message.MessageID)
		bot.Send(msg)
	}
}

func processUpdate(game *Game, update tgbotapi.Update) GameState {

	state := game.State

	switch state {
	case StateStart:
		if update.CallbackQuery != nil {
			if update.CallbackQuery.Data == "start_game" {
				return StateSelectTrump
			}
		}
	case StateSelectTrump:
		if update.CallbackQuery != nil {
			switch update.CallbackQuery.Data {
			case "red_team":
				game.Items[len(game.Items)-1].TrumpTeam = RedTeam
			case "black_team":
				game.Items[len(game.Items)-1].TrumpTeam = BlackTeam
			}
			return StateSelectHand
		}
	case StateSelectHand:
		if update.CallbackQuery != nil {
			value := update.CallbackQuery.Data
			game.Items[len(game.Items)-1].TrumpScore, _ = strconv.Atoi(value)
			return StateInputOtherScore
		}
	case StateInputOtherScore:
		if update.CallbackQuery != nil {
			value := update.CallbackQuery.Data
			game.Items[len(game.Items)-1].OpponentScore, _ = strconv.Atoi(value)
			return StateInputOtherScore
		}
	}

	return state
}

func sendResponse(state GameState, chatid int64, msgid int) tgbotapi.MessageConfig {
	switch state {
	case StateStart:
		return createStartGame(chatid, msgid)
	case StateSelectTrump:
		return createSelectTeam(chatid, msgid)
	case StateSelectHand:
		return createSelectHand(chatid, msgid)
	case StateInputOtherScore:
		return createSelectHand(chatid, msgid)
	}

	return createStartGameWithError(chatid, msgid)
}

func createStartGame(chatid int64, msgid int) tgbotapi.MessageConfig {
	replyKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("شروع بازی جدید", "start_game"),
		),
	)

	msg := tgbotapi.NewMessage(chatid, "به بات خوش آمدید!")
	msg.ReplyMarkup = replyKeyboard
	msg.ReplyToMessageID = msgid
	return msg
}

func createStartGameWithError(chatid int64, msgid int) tgbotapi.MessageConfig {
	replyKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("شروع بازی جدید", "start_game"),
		),
	)

	msg := tgbotapi.NewMessage(chatid, "دستور ورودی قابل پردازش نیست")
	msg.ReplyMarkup = replyKeyboard
	msg.ReplyToMessageID = msgid
	return msg
}

func createSelectTeam(chatid int64, msgid int) tgbotapi.MessageConfig {
	replyKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("تیم قرمز", "red_team"),
			tgbotapi.NewInlineKeyboardButtonData("تیم سیاه", "black_team"),
		),
	)

	msg := tgbotapi.NewMessage(chatid, "کدام تیم حاکم شد؟")
	msg.ReplyMarkup = replyKeyboard
	msg.ReplyToMessageID = msgid
	return msg
}

func createSelectHand(chatid int64, msgid int) tgbotapi.MessageConfig {
	replyKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("100", "100"),
			tgbotapi.NewInlineKeyboardButtonData("105", "105"),
			tgbotapi.NewInlineKeyboardButtonData("110", "110"),
			tgbotapi.NewInlineKeyboardButtonData("115", "115"),
			tgbotapi.NewInlineKeyboardButtonData("120", "120"),
			tgbotapi.NewInlineKeyboardButtonData("125", "125"),
			tgbotapi.NewInlineKeyboardButtonData("130", "130"),
			tgbotapi.NewInlineKeyboardButtonData("135", "135"),
			tgbotapi.NewInlineKeyboardButtonData("140", "140"),
			tgbotapi.NewInlineKeyboardButtonData("145", "145"),
			tgbotapi.NewInlineKeyboardButtonData("150", "150"),
			tgbotapi.NewInlineKeyboardButtonData("155", "155"),
			tgbotapi.NewInlineKeyboardButtonData("160", "160"),
			tgbotapi.NewInlineKeyboardButtonData("165", "شلم"),
			tgbotapi.NewInlineKeyboardButtonData("330", "سرشلم"),
		),
	)

	msg := tgbotapi.NewMessage(chatid, "حاکم چند خواند؟")
	msg.ReplyMarkup = replyKeyboard
	msg.ReplyToMessageID = msgid
	return msg
}

func calc(g *GameItem) (int, error) {
	if g.OpponentScore < 0 || g.OpponentScore > 160 || g.OpponentScore%5 != 0 {
		return 0, errors.New("یه خطا اتفاق افتاد")
	}

	fmt.Println(g)

	if g.OpponentScore == 0 {
		return 2 * g.Claim, nil
	} else if g.OpponentScore >= 85 {
		return -2 * g.Claim, nil
	}

	score := 165 - g.OpponentScore
	if score >= g.Claim {
		return g.Claim, nil
	} else {
		return -g.Claim, nil
	}
}

func showScore(bot *tgbotapi.BotAPI, chatID int64, game *Game) {
	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("امتیازات:\nقرمز: %d\nسیاه: %d", 100, 200))
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("دست جدید", "شروع بازی جدید"),
		),
	)
	bot.Send(msg)
}
