package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

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

	//bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)
	db := make(map[int64]*Game)

	for update := range updates {

		chat := update.FromChat()

		if chat == nil {
			continue
		}

		state, ok := db[chat.ID]

		if !ok {
			state = &Game{}
			db[chat.ID] = state
		}

		state.State = processUpdate(state, update)
		msg := sendResponse(state, chat.ID)
		bot.Send(msg)
	}
}

func processUpdate(game *Game, update tgbotapi.Update) GameState {

	state := game.State

	switch state {
	case StateStart:
		return StateNewGame
	case StateNewGame:
		return StateNewHand
	case StateNewHand:
		game.Items = append(game.Items, &GameItem{})
		return StateSelectTrump
	case StateSelectTrump:
		switch update.Message.Text {
		case "قرمز":
			game.Items[len(game.Items)-1].TrumpTeam = RedTeam
		case "سیاه":
			game.Items[len(game.Items)-1].TrumpTeam = BlackTeam
		}
		return StateSelectHand
	case StateSelectHand:
		value, err := strconv.Atoi(update.Message.Text)
		if err != nil {
			return StateSelectHand
		}
		game.Items[len(game.Items)-1].Claim = value
		return StateInputOtherScore
	case StateInputOtherScore:
		value, err := strconv.Atoi(update.Message.Text)
		if err != nil {
			return StateInputOtherScore
		}
		item := game.Items[len(game.Items)-1]
		item.OpponentScore = value
		score, err := calc(item)
		if err != nil {
			return StateInputOtherScore
		}
		item.TrumpScore = score

		updateTotal(game)

		return StateNewHand
	}

	return state
}

func sendResponse(game *Game, chatid int64) tgbotapi.MessageConfig {
	switch game.State {
	case StateNewGame:
		return createStartGame(chatid)
	case StateNewHand:
		return createStartHand(chatid, game)
	case StateSelectTrump:
		return createSelectTeam(chatid)
	case StateSelectHand:
		return createSelectHand(chatid)
	case StateInputOtherScore:
		return createSelectScore(chatid)
	}

	return createStartGameWithError(chatid)
}

func createStartGame(chatid int64) tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(chatid, "برای شروع بازی جدید کلیک کنید")
	msg.ReplyMarkup = tgbotapi.NewOneTimeReplyKeyboard(
		[]tgbotapi.KeyboardButton{
			tgbotapi.NewKeyboardButton("شروع بازی جدید"),
		},
	)
	return msg
}

func createStartHand(chatID int64, game *Game) tgbotapi.MessageConfig {
	text := fmt.Sprintf("امتیاز تیم قرمز: %d\nامتیاز تیم سیاه: %d", game.RedTeamScore, game.BalckTeamScore)
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = tgbotapi.NewOneTimeReplyKeyboard(
		[]tgbotapi.KeyboardButton{
			tgbotapi.NewKeyboardButton("ثبت دست جدید"),
		},
	)
	return msg
}

func createStartGameWithError(chatid int64) tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(chatid, "دستور ورودی قابل پردازش نیست")
	msg.ReplyMarkup = tgbotapi.NewOneTimeReplyKeyboard(
		[]tgbotapi.KeyboardButton{
			tgbotapi.NewKeyboardButton("شروع بازی جدید"),
		},
	)
	return msg
}

func createSelectTeam(chatid int64) tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(chatid, "کدام تیم حاکم است؟")
	msg.ReplyMarkup = tgbotapi.NewOneTimeReplyKeyboard(
		[]tgbotapi.KeyboardButton{
			tgbotapi.NewKeyboardButton("قرمز"),
			tgbotapi.NewKeyboardButton("سیاه"),
		},
	)
	return msg
}

func createSelectHand(chatid int64) tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(chatid, "حاکم چند خواند؟")
	msg.ReplyMarkup = tgbotapi.NewOneTimeReplyKeyboard(
		[]tgbotapi.KeyboardButton{
			tgbotapi.NewKeyboardButton("100"),
			tgbotapi.NewKeyboardButton("105"),
			tgbotapi.NewKeyboardButton("110"),
			tgbotapi.NewKeyboardButton("115"),
		},
		[]tgbotapi.KeyboardButton{
			tgbotapi.NewKeyboardButton("120"),
			tgbotapi.NewKeyboardButton("125"),
			tgbotapi.NewKeyboardButton("130"),
			tgbotapi.NewKeyboardButton("135"),
		},
		[]tgbotapi.KeyboardButton{
			tgbotapi.NewKeyboardButton("140"),
			tgbotapi.NewKeyboardButton("145"),
			tgbotapi.NewKeyboardButton("150"),
			tgbotapi.NewKeyboardButton("155"),
		},
		[]tgbotapi.KeyboardButton{
			tgbotapi.NewKeyboardButton("160"),
			tgbotapi.NewKeyboardButton("شلم"),
			tgbotapi.NewKeyboardButton("سرشلم"),
		},
	)
	return msg
}

func createSelectScore(chatID int64) tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(chatID, "حریف حاکم چند امتیاز گرفت؟")
	msg.ReplyMarkup = tgbotapi.NewOneTimeReplyKeyboard(
		[]tgbotapi.KeyboardButton{
			tgbotapi.NewKeyboardButton("0"),
			tgbotapi.NewKeyboardButton("5"),
			tgbotapi.NewKeyboardButton("10"),
			tgbotapi.NewKeyboardButton("15"),
			tgbotapi.NewKeyboardButton("20"),
			tgbotapi.NewKeyboardButton("25"),
			tgbotapi.NewKeyboardButton("30"),
			tgbotapi.NewKeyboardButton("35"),
		},
		[]tgbotapi.KeyboardButton{
			tgbotapi.NewKeyboardButton("40"),
			tgbotapi.NewKeyboardButton("45"),
			tgbotapi.NewKeyboardButton("50"),
			tgbotapi.NewKeyboardButton("55"),
			tgbotapi.NewKeyboardButton("60"),
			tgbotapi.NewKeyboardButton("65"),
			tgbotapi.NewKeyboardButton("70"),
			tgbotapi.NewKeyboardButton("75"),
			tgbotapi.NewKeyboardButton("80"),
		},
	)
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
	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("امتیازات:\nقرمز: %d\nسیاه: %d", game.RedTeamScore, game.BalckTeamScore))
	bot.Send(msg)
}

func updateTotal(game *Game) {

	item := game.Items[len(game.Items)-1]

	if item.TrumpTeam == BlackTeam {
		game.BalckTeamScore += item.TrumpScore
		game.RedTeamScore += item.OpponentScore
	} else {
		game.RedTeamScore += item.TrumpScore
		game.BalckTeamScore += item.OpponentScore
	}

}

func debugProxy() {
	req, _ := http.NewRequest("GET", "https://google.com", nil)
	p, _ := http.ProxyFromEnvironment(req)
	log.Println("USING PROXY:", p)
}
