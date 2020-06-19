package welcome

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/PaulSonOfLars/gotgbot"
	"github.com/PaulSonOfLars/gotgbot/ext"
	"github.com/anandpskerala/Martha/bot/modules/sql"
	"github.com/anandpskerala/Martha/bot/modules/utils/chats"
	"github.com/anandpskerala/Martha/bot/modules/utils/helpers"
)

func welcome(bot ext.Bot, u *gotgbot.Update, args []string) error {
	chat := u.EffectiveChat
	vchatId := u.EffectiveChat.Id
	vchat := sql.GetChatConnect(strconv.Itoa(vchatId))

	if u.EffectiveChat.Type == "private" {
		if vchat.ChatId != "" {
			vc, _ := strconv.Atoi(vchat.ChatId)
			chat, _ = bot.GetChat(vc)
		} else {
			_, err := u.EffectiveMessage.ReplyText("This command is meant to be used in a group!")
			return err
		}
	}

	if !chats.IsUserAdmin(chat, u.EffectiveUser.Id) {
		_, _ = u.EffectiveMessage.ReplyText("You need to be an admin to do this.")
		return gotgbot.ContinueGroups{}
	}

	if len(args) == 0 || strings.ToLower(args[0]) == "noformat" {
		noformat := len(args) > 0 && strings.ToLower(args[0]) == "noformat"
		welcPrefs := sql.GetWelcomePrefs(strconv.Itoa(chat.Id))
		_, _ = u.EffectiveMessage.ReplyHTMLf("Welcome settings in <b>%v</b>:\nI am currently welcoming users: <code>%v</code>"+
			"\nI am currently deleting old welcomes: <code>%v</code>"+
			"\nI am currently deleting service messages: <code>%v</code>"+
			"\nOn joining, I am currently muting users: <code>%v</code>"+
			"\nThe welcome message not filling the {} is:",
			chat.Title,
			welcPrefs.ShouldWelcome,
			welcPrefs.CleanWelcome != 0,
			welcPrefs.DelJoined,
			welcPrefs.ShouldMute)

		if welcPrefs.WelcomeType == sql.BUTTON_TEXT {
			buttons := sql.GetWelcomeButtons(strconv.Itoa(chat.Id))
			if strings.Contains(welcPrefs.CustomWelcome, "{rules}") {
				rulesButton := sql.WelcomeButton{
					Id:       0,
					ChatId:   strconv.Itoa(u.EffectiveChat.Id),
					Name:     "Rules",
					Url:      fmt.Sprintf("t.me/%v?start=%v", bot.UserName, u.EffectiveChat.Id),
					SameLine: false,
				}
				buttons = append(buttons, rulesButton)
				strings.ReplaceAll(welcPrefs.CustomWelcome, "{rules}", "")
			}
			if noformat {
				welcPrefs.CustomWelcome += helpers.RevertButtons(buttons)
				_, err := u.EffectiveMessage.ReplyHTML(welcPrefs.CustomWelcome)
				return err
			} else {
				keyb := helpers.BuildWelcomeKeyboard(buttons)
				keyboard := ext.InlineKeyboardMarkup{InlineKeyboard: &keyb}
				send(bot, u, welcPrefs.CustomWelcome, &keyboard, sql.DefaultWelcome, !welcPrefs.DelJoined)
			}
		} else {
			_, err := EnumFuncMap[welcPrefs.WelcomeType](bot, chat.Id, welcPrefs.CustomWelcome) // needs change
			return err
		}
	} else if len(args) >= 1 {
		switch strings.ToLower(args[0]) {
		case "on", "yes":
			go sql.SetWelcPref(strconv.Itoa(chat.Id), true)
			_, err := u.EffectiveMessage.ReplyText("I'll welcome users from now on.")
			return err
		case "off", "no":
			go sql.SetWelcPref(strconv.Itoa(chat.Id), false)
			_, err := u.EffectiveMessage.ReplyText("I'll not welcome users from now on.")
			return err
		default:
			_, err := u.EffectiveMessage.ReplyText("I understand 'on/yes' or 'off/no' only!")
			return err
		}
	}
	return nil
}

func setWelcome(b ext.Bot, u *gotgbot.Update) error {
	chat := u.EffectiveChat
	msg := u.EffectiveMessage
	vchatId := u.EffectiveChat.Id
	vchat := sql.GetChatConnect(strconv.Itoa(vchatId))

	if u.EffectiveChat.Type == "private" {
		if vchat.ChatId != "" {
			vc, _ := strconv.Atoi(vchat.ChatId)
			chat, _ = b.GetChat(vc)
		} else {
			_, err := u.EffectiveMessage.ReplyText("This command is meant to be used in a group!")
			return err
		}
	}

	if !chats.IsUserAdmin(chat, u.EffectiveUser.Id) {
		_, _ = u.EffectiveMessage.ReplyText("You need to be an admin to do this.")
		return gotgbot.ContinueGroups{}
	}

	text, dataType, content, buttons := helpers.GetWelcomeType(msg)
	if dataType == -1 {
		_, err := msg.ReplyText("You didn't specify what to reply with!")
		return err
	}

	btns := make([]sql.WelcomeButton, len(buttons))
	for i, btn := range buttons {
		btns[i] = sql.WelcomeButton{
			ChatId:   strconv.Itoa(chat.Id),
			Name:     btn.Name,
			Url:      btn.Content,
			SameLine: btn.SameLine,
		}
	}

	if text != "" {
		go sql.SetCustomWelcome(strconv.Itoa(chat.Id), text, btns, dataType)
	} else {
		go sql.SetCustomWelcome(strconv.Itoa(chat.Id), content, btns, dataType)
	}

	_, err := msg.ReplyTextf("Successfully set custom welcome message in <b>%v</b>!", chat.Title)
	return err
}

func resetWelcome(b ext.Bot, u *gotgbot.Update) error {
	chat := u.EffectiveChat

	vchatId := u.EffectiveChat.Id
	vchat := sql.GetChatConnect(strconv.Itoa(vchatId))

	if u.EffectiveChat.Type == "private" {
		if vchat.ChatId != "" {
			vc, _ := strconv.Atoi(vchat.ChatId)
			chat, _ = b.GetChat(vc)
		} else {
			_, err := u.EffectiveMessage.ReplyText("This command is meant to be used in a group!")
			return err
		}
	}

	if !chats.IsUserAdmin(chat, u.EffectiveUser.Id) {
		_, _ = u.EffectiveMessage.ReplyText("You need to be an admin to do this.")
		return gotgbot.ContinueGroups{}
	}

	go sql.SetCustomWelcome(strconv.Itoa(chat.Id), sql.DefaultWelcome, nil, sql.TEXT)

	_, err := u.EffectiveMessage.ReplyTextf("Succesfully reset custom welcome message to default in <b>5v</b>!", chat.Title)
	return err
}

func cleanWelcome(b ext.Bot, u *gotgbot.Update, args []string) error {
	chat := u.EffectiveChat
	vchatId := u.EffectiveChat.Id
	vchat := sql.GetChatConnect(strconv.Itoa(vchatId))

	if u.EffectiveChat.Type == "private" {
		if vchat.ChatId != "" {
			vc, _ := strconv.Atoi(vchat.ChatId)
			chat, _ = b.GetChat(vc)
		} else {
			_, err := u.EffectiveMessage.ReplyText("This command is meant to be used in a group!")
			return err
		}
	}

	if !chats.IsUserAdmin(chat, u.EffectiveUser.Id) {
		_, _ = u.EffectiveMessage.ReplyText("You need to be an admin to do this.")
		return gotgbot.ContinueGroups{}
	}

	if len(args) == 0 {
		cleanPref := sql.GetCleanWelcome(strconv.Itoa(chat.Id))
		if cleanPref != 0 {
			_, err := u.EffectiveMessage.ReplyText("I should be deleting welcome messages up to two days old.")
			return err
		} else {
			_, err := u.EffectiveMessage.ReplyText("I'm currently not deleting old welcome messages!")
			return err
		}
	}

	switch strings.ToLower(args[0]) {
	case "off", "no":
		go sql.SetCleanWelcome(strconv.Itoa(chat.Id), 0)
		_, err := u.EffectiveMessage.ReplyText("I'll try to delete old welcome messages!")
		return err
	case "on", "yes":
		go sql.SetCleanWelcome(strconv.Itoa(chat.Id), 1)
		_, err := u.EffectiveMessage.ReplyText("I'll try to delete old welcome messages!")
		return err
	default:
		_, err := u.EffectiveMessage.ReplyText("I understand 'on/yes' or 'off/no' only!")
		return err
	}
}

func delJoined(b ext.Bot, u *gotgbot.Update, args []string) error {
	chat := u.EffectiveChat

	vchatId := u.EffectiveChat.Id
	vchat := sql.GetChatConnect(strconv.Itoa(vchatId))

	if u.EffectiveChat.Type == "private" {
		if vchat.ChatId != "" {
			vc, _ := strconv.Atoi(vchat.ChatId)
			chat, _ = b.GetChat(vc)
		} else {
			_, err := u.EffectiveMessage.ReplyText("This command is meant to be used in a group!")
			return err
		}
	}
	if !chats.IsUserAdmin(chat, u.EffectiveUser.Id) {
		_, _ = u.EffectiveMessage.ReplyText("You need to be an admin to do this.")
		return gotgbot.ContinueGroups{}
	}

	if len(args) == 0 {
		delPref := sql.GetDelPref(strconv.Itoa(chat.Id))
		if delPref {
			_, err := u.EffectiveMessage.ReplyMarkdown("I should be deleting `user` joined the chat messages now.")
			return err
		} else {
			_, err := u.EffectiveMessage.ReplyText("I'm currently not deleting joined messages.")
			return err
		}
	}

	switch strings.ToLower(args[0]) {
	case "off", "no":
		go sql.SetDelPref(strconv.Itoa(chat.Id), false)
		_, err := u.EffectiveMessage.ReplyText("I won't delete joined messages.")
		return err
	case "on", "yes":
		go sql.SetDelPref(strconv.Itoa(chat.Id), true)
		_, err := u.EffectiveMessage.ReplyText("I'll try to delete joined messages!")
		return err
	default:
		_, err := u.EffectiveMessage.ReplyText("I understand 'on/yes' or 'off/no' only!")
		return err
	}
}

func welcomeMute(b ext.Bot, u *gotgbot.Update, args []string) error {
	chat := u.EffectiveChat

	vchatId := u.EffectiveChat.Id
	vchat := sql.GetChatConnect(strconv.Itoa(vchatId))

	if u.EffectiveChat.Type == "private" {
		if vchat.ChatId != "" {
			vc, _ := strconv.Atoi(vchat.ChatId)
			chat, _ = b.GetChat(vc)
		} else {
			_, err := u.EffectiveMessage.ReplyText("This command is meant to be used in a group!")
			return err
		}
	}

	if !chats.IsUserAdmin(chat, u.EffectiveUser.Id) {
		_, _ = u.EffectiveMessage.ReplyText("You need to be an admin to do this.")
		return gotgbot.ContinueGroups{}
	}

	if len(args) == 0 {
		welcPref := sql.GetWelcomePrefs(strconv.Itoa(chat.Id))
		if welcPref.ShouldMute {
			_, err := u.EffectiveMessage.ReplyMarkdown("I'm currently muting users when they join.")
			return err
		} else {
			_, err := u.EffectiveMessage.ReplyText("I'm currently not muting users when they join.")
			return err
		}
	}

	switch strings.ToLower(args[0]) {
	case "off", "no":
		go sql.SetMutePref(strconv.Itoa(chat.Id), false)
		_, err := u.EffectiveMessage.ReplyText("I won't mute new users when they join.")
		return err
	case "on", "yes":
		go sql.SetMutePref(strconv.Itoa(chat.Id), true)
		_, err := u.EffectiveMessage.ReplyText("I'll try to mute new users when they join!")
		return err
	default:
		_, err := u.EffectiveMessage.ReplyText("I understand 'on/yes' or 'off/no' only!")
		return err
	}
}
