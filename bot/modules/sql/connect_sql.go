package sql

import (
	"encoding/json"
	"fmt"

	"github.com/wI2L/jettison"

	"github.com/anandpskerala/Martha/bot/modules/utils/caching"
        "github.com/anandpskerala/Martha/bot/modules/utils/error_handling"
)

type Connect struct {
	UserId string `gorm:"primary_key" json:"user_id"`
	ChatId  string `json:"connect"`
	ChatTitle string
}


func SetChatConnect(userId, chatId, chatTitle string) {
	SESSION.Save(&Connect{UserId: userId, ChatId: chatId, ChatTitle: chatTitle})
	cacheConnect(userId)
}

func GetChatConnect(userId string) *Connect {
	connectJson, err := caching.CACHE.Get(fmt.Sprintf("connect_%v", userId))
	var connect *Connect
	if err != nil {
		connect = cacheConnect(userId)
		connectJson, err = caching.CACHE.Get(fmt.Sprintf("connect_%v", userId))
		if err != nil {
			connect = cacheConnect(userId)
		}
		
	}
	_ = json.Unmarshal(connectJson, &connect)
	return connect
}


func cacheConnect(userId string) *Connect {
	connect := &Connect{}
	SESSION.Where("user_id = ?", userId).Find(&connect)
	connectJson, _ := jettison.Marshal(&connect)
	err := caching.CACHE.Set(fmt.Sprintf("connect_%v", userId), connectJson)
	error_handling.HandleErr(err)
	return connect
}
