package sql

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/anandpskerala/Martha/bot/modules/utils/caching"
	"github.com/anandpskerala/Martha/bot/modules/utils/error_handling"
	"github.com/PaulSonOfLars/gotgbot"
)

type User struct {
	UserId   int    `gorm:"primary_key" json:"user_id"`
	UserName string `json:"user_name"`
}

type Chat struct {
	ChatId   string `gorm:"primary_key" json:"chat_id"`
	ChatName string `json:"chat_name"`
}

func EnsureBotInDb(u *gotgbot.Updater) {
	// Insert bot user only if it doesn't exist already
	botUser := &User{UserId: u.Dispatcher.Bot.Id, UserName: u.Dispatcher.Bot.UserName}
	SESSION.Save(botUser)
}

func UpdateUser(userId int, username string, chatId string, chatName string) {
	username = strings.ToLower(username)
	defer func(name string) {
		go cacheUser()
	}(username)
	tx := SESSION.Begin()

	// upsert user
	user := &User{UserId: userId, UserName: username}
	tx.Save(user)

	if chatId == "nil" || chatName == "nil" {
		tx.Commit()
		return
	}

	// upsert chat
	chat := &Chat{ChatId: chatId, ChatName: chatName}
	tx.Save(chat)
	tx.Commit()
}

func GetUserIdByName(username string) *User {
	username = strings.ToLower(username)

	userJson, err := caching.CACHE.Get("users")
	if err != nil {
		go cacheUser()
		userJson, err = caching.CACHE.Get("users")
		if err != nil {
			go cacheUser()
		}
	}

	log.Println(string(userJson))

	var users []User
	_ = json.Unmarshal(userJson, &users)
	for _, user := range users {
		if user.UserName == username {
			return &user
		}
	}

	return nil
}

func cacheUser() {
	user := &User{}
	var users []User
	SESSION.Model(user).Find(&users)
	userJson, _ := json.Marshal(users)
	err := caching.CACHE.Set("users", userJson)
	error_handling.HandleErr(err)
}
