package sql

import (
	"encoding/json"
	"fmt"

	"github.com/wI2L/jettison"

	"github.com/anandpskerala/Martha/bot/modules/utils/caching"
)

type BlackListFilters struct {
	ChatID  string `gorm:"primary_key" json:"chat_id"`
	Trigger string `gorm:"primary_key" json:"trigger"`
}

func AddToBlacklist(chatID string, trigger string) {
	filter := &BlackListFilters{ChatID: chatID, Trigger: trigger}
	SESSION.Save(filter)
	cacheBlacklist(chatID)
}

func RmFromBlacklist(chatID string, trigger string) bool {
	filter := &BlackListFilters{ChatID: chatID, Trigger: trigger}
	if SESSION.Delete(filter).RowsAffected == 0 {
		return false
	}
	cacheBlacklist(chatID)
	return true
}

func GetChatBlacklist(chatID string) []BlackListFilters {
	blf, err := caching.CACHE.Get(fmt.Sprintf("blacklist_%v", chatID))
	var blistFilters []BlackListFilters = nil
	if err != nil {
		blistFilters = cacheBlacklist(chatID)
		blf, err = caching.CACHE.Get(fmt.Sprintf("blacklist_%v", chatID))
		if err != nil {
			blistFilters = cacheBlacklist(chatID)
		}
	}

	_ = json.Unmarshal(blf, &blistFilters)
	return blistFilters
}

func cacheBlacklist(chatID string) []BlackListFilters {
	var filters []BlackListFilters
	SESSION.Where("chat_id = ?", chatID).Find(&filters)
	blJson, _ := jettison.Marshal(filters)
	_ = caching.CACHE.Set(fmt.Sprintf("blacklist_%v", chatID), blJson)
	return filters
}
