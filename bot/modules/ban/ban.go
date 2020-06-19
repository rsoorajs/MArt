package bans

import (
	"log"
	"strings"

	"github.com/anandpskerala/Martha/bot/modules/utils/chats"
	"github.com/anandpskerala/Martha/bot/modules/utils/error_handling"
	"github.com/anandpskerala/Martha/bot/modules/utils/extraction"
       // "github.com/anandpskerala/Martha/bot/modules/utils/helpers"
	"github.com/anandpskerala/Martha/bot/modules/utils/string_handling"
	"github.com/PaulSonOfLars/gotgbot"
	"github.com/PaulSonOfLars/gotgbot/ext"
	"github.com/PaulSonOfLars/gotgbot/handlers"
)

func ban(bot ext.Bot, u *gotgbot.Update, args []string) error {
	chat := u.EffectiveChat
	user := u.EffectiveUser
	message := u.EffectiveMessage
        tuser := u.EffectiveMessage.ReplyToMessage.From
        

	// Permission checks
	if u.EffectiveChat.Type == "private" {
		_, err := u.EffectiveMessage.ReplyText("This command is meant to be used in a group!")
		return err
	}

	if !chats.RequireBotAdmin(chat, message) {
		return gotgbot.EndGroups{}
	}
	if !chats.RequireUserAdmin(chat, message, user.Id) {
		return gotgbot.EndGroups{}
	}

	userId, _ := extraction.ExtractUserAndText(message, args)
	if userId == 0 {
		_, err := message.ReplyText("Try targeting a user buddy.")
		return err
	}

	member, err := chat.GetMember(userId)
	if err != nil {
		if err.Error() == "User not found" {
			_, err = message.ReplyText("This user is dead mate.")
		}
		return err
	}

	userMember, _ := chat.GetMember(user.Id)
	if !userMember.CanRestrictMembers && userMember.Status != "creator" {
		_, err = message.ReplyText("You don't have permissions to ban users!")
		return err
	}

	if chats.IsUserBanProtected(chat, userId, member) {
		_, err := message.ReplyText("One day I'll find out how to work around the bot API. Today is not that day.")
		return err
	}

	if userId == bot.Id {
		_, err := message.ReplyText("No you")
		return err
	}

	_, err = chat.KickMember(userId)
	if err != nil {
		return err
	}
        if message.ReplyToMessage != nil {
                _, err = message.ReplyHTMLf("Another bit of dust..\nBanned <a href='tg://user?id=%d'>%s</a>!", tuser.Id, tuser.FirstName)
                return err
        }// else {
        _, err = message.ReplyHTMLf("Banned!")
        return err
       // }

	//_, err = message.ReplyHTMLf(banstring)
	//return err
}

func tempBan(bot ext.Bot, u *gotgbot.Update, args []string) error {
	chat := u.EffectiveChat
	user := u.EffectiveUser
	message := u.EffectiveMessage
        tuser := u.EffectiveMessage.ReplyToMessage.From
        

	// Permission checks
	if u.EffectiveChat.Type == "private" {
		_, err := u.EffectiveMessage.ReplyText("This command is meant to be used in a group!")
		return err
	}

	if !chats.RequireBotAdmin(chat, message) {
		return gotgbot.EndGroups{}
	}
	if !chats.RequireUserAdmin(chat, message, user.Id) {
		return gotgbot.EndGroups{}
	}

	userId, reason := extraction.ExtractUserAndText(message, args)
	if userId == 0 {
		_, err := message.ReplyText("Try targeting a user buddy.")
		return err
	}

	member, err := chat.GetMember(userId)
	if err != nil {
		if err.Error() == "User not found" {
			_, err = message.ReplyText("This user is dead mate.")
		}
		return err
	}

	userMember, _ := chat.GetMember(user.Id)
	if !userMember.CanRestrictMembers && userMember.Status != "creator" {
		_, err = message.ReplyText("You don't have permissions to ban users!")
		return err
	}

	if chats.IsUserBanProtected(chat, userId, member) {
		_, err := message.ReplyText("One day I'll find out how to work around the bot API. Today is not that day.")
		return err
	}

	if userId == bot.Id {
		_, err := message.ReplyText("No you can't")
		return err
	}

	if reason == "" {
		_, err := message.ReplyText("I don't know how long I'm supposed to ban them for 🤔.")
		return err
	}

	splitReason := strings.SplitN(reason, " ", 2)
	timeVal := splitReason[0]
	banTime := string_handling.ExtractTime(message, timeVal)
	if banTime == -1 {
		return nil
	}
	newMsg := bot.NewSendableKickChatMember(chat.Id, userId)
	string_handling.ExtractTime(message, timeVal)
	newMsg.UntilDate = banTime
	_, err = newMsg.Send()
	if err != nil {
		_, err := message.ReplyText("I can't ban this user.")
		error_handling.HandleErr(err)
	}
        if message.ReplyToMessage != nil {
                _, err = message.ReplyHTMLf("Banned <a href='tg://user?id=%d'>%s</a> for <b>%s</b>!", tuser.Id, tuser.FirstName, timeVal)
                return err
        } //else {
        _, err = message.ReplyHTMLf("Banned for <b>%s</b>!", timeVal)
        return err
        //}
	//_, err = message.ReplyHTMLf(tbanstrng)
	//return err
}

func kick(bot ext.Bot, u *gotgbot.Update, args []string) error {
	chat := u.EffectiveChat
	user := u.EffectiveUser
	message := u.EffectiveMessage
        tuser := u.EffectiveMessage.ReplyToMessage.From
        

	// Permission checks
	if u.EffectiveChat.Type == "private" {
		_, err := u.EffectiveMessage.ReplyText("This command is meant to be used in a group!")
		return err
	}

	if !chats.RequireBotAdmin(chat, message) {
		return gotgbot.EndGroups{}
	}
	if !chats.RequireUserAdmin(chat, message, user.Id) {
		return gotgbot.EndGroups{}
	}

	userId, _ := extraction.ExtractUserAndText(message, args)
	if userId == 0 {
		_, err := message.ReplyText("Try targeting a user buddy.")
		return err
	}

	var member, err = chat.GetMember(userId)
	if err != nil {
		if err.Error() == "User not found" {
			_, err = message.ReplyText("This user is dead mate.")
		}
		return err
	}

	userMember, _ := chat.GetMember(user.Id)
	if !userMember.CanRestrictMembers && userMember.Status != "creator" {
		_, err = message.ReplyText("You don't have permissions to ban users!")
		return err
	}

	if chats.IsUserBanProtected(chat, userId, member) {
		_, err := message.ReplyText("One day I'll find out how to work around the bot API. Today is not that day.")
		return err
	}

	if userId == bot.Id {
		_, err := message.ReplyText("No u can't")
		return err
	}

	_, err = chat.UnbanMember(userId) // Apparently unban on current user = kick
	if err != nil {
		_, err = message.ReplyText("Hec, I can't seem to kick this user.")
		return err
	}
        if message.ReplyToMessage != nil {
                _, err = message.ReplyHTMLf("Kicked <b>%s</b> !", tuser.FirstName)
                return err
        } //else {
        _, err = message.ReplyHTMLf("<b>Kicked!</b>")
        return err
        //}
	//_, err = message.ReplyHTMLf(kickstring)
	//return err
}

func kickme(_ ext.Bot, u *gotgbot.Update) error {
	chat := u.EffectiveChat
	user := u.EffectiveUser
	message := u.EffectiveMessage

	// Permission checks
	if u.EffectiveChat.Type == "private" {
		_, err := u.EffectiveMessage.ReplyText("This command is meant to be used in a group!")
		return err
	}

	if !chats.RequireBotAdmin(chat, message) {
		return gotgbot.EndGroups{}
	}

	if chats.IsUserAdmin(chat, user.Id) {
		_, err := message.ReplyText("Admin sir pls ;_;")
		error_handling.HandleErr(err)
		return gotgbot.EndGroups{}
	}
	bb, _ := chat.UnbanMember(user.Id)
	if bb {
		_, err := message.ReplyText("Ok No problem.")
		return err
	} else {
		_, err := message.ReplyText("Oh no I can't :/")
		return err
	}
}

func unban(bot ext.Bot, u *gotgbot.Update, args []string) error {
	chat := u.EffectiveChat
	user := u.EffectiveUser
	message := u.EffectiveMessage

	// Permission checks
	if u.EffectiveChat.Type == "private" {
		_, err := u.EffectiveMessage.ReplyText("This command is meant to be used in a group!")
		return err
	}

	if !chats.RequireBotAdmin(chat, message) && chats.RequireUserAdmin(chat, message, user.Id) {
		return gotgbot.EndGroups{}
	}

	userId, _ := extraction.ExtractUserAndText(message, args)

	if userId == 0 {
		_, err := message.ReplyText("Try targeting a user next time buddy.")
		return err
	}

	_, err := chat.GetMember(userId)
	if err != nil {
		_, err := message.ReplyText("This user is dead mate.")
		return err
	}

	userMember, _ := chat.GetMember(user.Id)
	if !userMember.CanRestrictMembers && userMember.Status != "creator" {
		_, err = message.ReplyText("You don't have permissions to unban users!")
		return err
	}

	if userId == bot.Id {
		_, err := message.ReplyText("What exactly are you trying to do?.")
		return err
	}

	if chats.IsUserInChat(chat, userId) {
		_, err := message.ReplyText("This user is already in the group!")
		return err
	}

	_, err = chat.UnbanMember(userId)
	error_handling.HandleErr(err)
	_, err = message.ReplyText("Fine, I'll allow him, for this time...")
	return err
}

func LoadBans(u *gotgbot.Updater) {
	defer log.Println("Loading module bans")
	u.Dispatcher.AddHandler(handlers.NewPrefixArgsCommand("tban", []rune{'/', '!'}, tempBan))
	u.Dispatcher.AddHandler(handlers.NewPrefixArgsCommand("ban", []rune{'/', '!'}, ban))
	u.Dispatcher.AddHandler(handlers.NewPrefixArgsCommand("kick", []rune{'/', '!'}, kick))
	u.Dispatcher.AddHandler(handlers.NewPrefixCommand("kickme", []rune{'/', '!'}, kickme))
	u.Dispatcher.AddHandler(handlers.NewPrefixArgsCommand("unban", []rune{'/', '!'}, unban))
}

