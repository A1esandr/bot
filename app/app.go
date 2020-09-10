package app

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

var (
	//ReaderFile define for test
	ReaderFile = ioutil.ReadFile
)

type App struct {
	token  string
	config *Config
	users  map[string][]User
}

type Config struct {
	Welcome string `json:"welcome"`
}

type User struct {
	Name    string
	Surname string
	Data    string
}

func NewApp() *App {
	return &App{config: &Config{}, users: map[string][]User{}}
}

func (a *App) init() {
	a.token = os.Getenv("TOKEN")
	if a.token == "" {
		log.Panic("token is empty!")
	}
	err := a.config.loadConfig()
	if err != nil {
		log.Panic(err.Error())
	}
	err = a.loadUsers()
	if err != nil {
		log.Panic(err.Error())
	}
}

func (c *Config) loadConfig() error {
	fileName := "../config/custom.json"
	data, err := ReaderFile(fileName)
	if err != nil {
		data, err = ReaderFile("../config/config.json")
		if err != nil {
			return err
		}
	}
	if err := json.Unmarshal(data, c); err != nil {
		return fmt.Errorf("Error to parse config %s", err)
	}
	return nil
}

func (a *App) loadUsers() error {
	data, err := ReaderFile("../config/users.csv")
	if err != nil {
		return err
	}
	r := csv.NewReader(strings.NewReader(string(data)))
	records, err := r.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	for _, record := range records {
		user := User{Name: record[1], Surname: record[0], Data: record[2]}
		if _, ok := a.users[user.Surname]; !ok {
			a.users[user.Surname] = []User{}
		}
		a.users[user.Surname] = append(a.users[user.Surname], user)
	}
	return nil
}

func (a *App) Start() {
	a.init()
	bot, err := tgbotapi.NewBotAPI(a.token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		msg.ReplyToMessageID = update.Message.MessageID

		bot.Send(msg)
	}
}

func (a *App) chooseMsg(command string) string {
	switch command {
	case "/start":
		return a.config.Welcome
	default:
		return ""
	}
}
