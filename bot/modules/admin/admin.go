package admin

import (
	"fmt"
	"html"
	"log"
	"strconv"
	//"strings"

	"github.com/anandpskerala/Martha/bot/modules/utils/chats"
	"github.com/anandpskerala/Martha/bot/modules/utils/error_handling"
	"github.com/anandpskerala/Martha/bot/modules/utils/extraction"
	"github.com/anandpskerala/Martha/bot/modules/utils/helpers"
	"github.com/anandpskerala/Martha/bot/modules/utils/string_handling"
	"github.com/PaulSonOfLars/gotgbot"
	"github.com/PaulSonOfLars/gotgbot/ext"
	"github.com/PaulSonOfLars/gotgbot/handlers"
)

func promote(bot ext.Bot, u *gotgbot.Update, args []string) error {
	chat := u.EffectiveChat
	chatId := chat.Id
	message := u.EffectiveMessage
	user := u.EffectiveUser

	// permission checks
	if chat.Type == "private" {
		_, err := message.ReplyText("This command is meant to be used in a group!")
		return err
	}
	if !chats.RequireBotAdmin(chat, message) {
		return gotgbot.EndGroups{}
	}
	if !chats.RequireUserAdmin(chat, message, user.Id) {
		return gotgbot.EndGroups{}
	}
	if !chats.CanPromote(bot, chat) {
		return gotgbot.EndGroups{}
	}

	userId := extraction.ExtractUser(message, args)
	if userId == 0 {
		_, err := message.ReplyText("This user is ded mate.")
		error_handling.HandleErr(err)
		return nil
	}

	userMember, err := chat.GetMember(userId)
	error_handling.HandleErr(err)

	if userId == bot.Id {
		_, err := message.ReplyText("If only I could do this to myself ;_;")
		error_handling.HandleErr(err)
		return nil
	}

	botMember, err := chat.GetMember(bot.Id)
	error_handling.HandleErr(err)

	sendablePromoteChatMember := bot.NewSendablePromoteChatMember(chatId, userId)
	sendablePromoteChatMember.CanDeleteMessages = botMember.CanDeleteMessages
	sendablePromoteChatMember.CanChangeInfo = botMember.CanDeleteMessages
	sendablePromoteChatMember.CanEditMessages = botMember.CanEditMessages
	sendablePromoteChatMember.CanPostMessages = botMember.CanPostMessages
	sendablePromoteChatMember.CanInviteUsers = botMember.CanInviteUsers
	sendablePromoteChatMember.CanPinMessages = botMember.CanPinMessages
	sendablePromoteChatMember.CanRestrictMembers = botMember.CanRestrictMembers
	sendablePromoteChatMember.CanPromoteMembers = botMember.CanPromoteMembers

	_, err = sendablePromoteChatMember.Send()
	error_handling.HandleErr(err)

	_, err = message.ReplyHTMLf("Successfully promoted %v!", helpers.MentionHtml(userId, userMember.User.FirstName))

	return err
}

func demote(bot ext.Bot, u *gotgbot.Update, args []string) error {
	chat := u.EffectiveChat
	chatId := chat.Id
	message := u.EffectiveMessage
	user := u.EffectiveUser

	// permission checks
	if chat.Type == "private" {
		_, err := message.ReplyText("This command is meant to be used in a group!")
		return err
	}
	if !chats.RequireBotAdmin(chat, message) {
		return gotgbot.EndGroups{}
	}
	if !chats.RequireUserAdmin(chat, message, user.Id) {
		return gotgbot.EndGroups{}
	}
	if !chats.CanPromote(bot, chat) {
		return gotgbot.EndGroups{}
	}

	userId := extraction.ExtractUser(message, args)
	if userId == 0 {
		_, err := message.ReplyText("This user is dead mate.")
		return err
	}

	userMember, err := chat.GetMember(userId)
	error_handling.HandleErr(err)

	if userMember.Status == "creator" {
		_, err := message.ReplyText("This person CREATED the chat, how could I demote them?")
		return err
	}

	if userId == bot.Id {
		_, err := message.ReplyText("Pls no sir ;_;")
		return err
	}

	bb, err := bot.DemoteChatMember(chatId, userId)
	if err != nil || !bb {
		log.Println(err)
		_, err := message.ReplyText("Could not demote. I might not be admin, or the admin status was appointed by another user, so I can't act upon them!")
		return err
	}

	_, err = message.ReplyText("Successfully demoted!")
	return err
}

func invitelink(bot ext.Bot, u *gotgbot.Update) error {
	user := u.EffectiveUser
	chat := u.EffectiveChat
	message := u.EffectiveMessage

	// Check permissions
	if chat.Type == "private" {
		_, err := message.ReplyText("This command is meant to be used in a group!")
		return err
	}
	if !chats.RequireUserAdmin(chat, message, user.Id) {
		return gotgbot.EndGroups{}
	}
	if !chats.RequireBotAdmin(chat, message) {
		return gotgbot.EndGroups{}
	}

	if chat.Username != "" {
		_, err := message.ReplyText(chat.Username)
		return err
	} else if chat.Type == "supergroup" || chat.Type == "channel" {
		botMember, err := chat.GetMember(bot.Id)
		error_handling.HandleErr(err)
		if botMember.CanInviteUsers {
			inviteLink, err := bot.ExportChatInviteLink(chat.Id)
			error_handling.HandleErr(err)
			_, err = message.ReplyText(inviteLink)
			return err
		} else {
			_, err := message.ReplyText("I don't have access to the invite link, try changing my permissions!")
			return err
		}
	} else {
		_, err := message.ReplyText("I can only give you invite links for supergroups and channels, sorry!")
		return err
	}
}

func adminlist(_ ext.Bot, u *gotgbot.Update) error {
	if u.EffectiveChat.Type == "private" {
		_, err := u.EffectiveMessage.ReplyText("This command is meant to be used in a group!")
		return err
	}

	admins, err := u.EffectiveChat.GetAdministrators()
	error_handling.HandleErr(err)
	var addendum string
	if u.EffectiveChat.Title != "" {
		addendum = u.EffectiveChat.Title
	} else {
		addendum = "This chat"
	}
	text := fmt.Sprintf("Admins in <b>%s</b>:", addendum)
	for _, admin := range admins {
		user := admin.User
		name := string_handling.FormatText("[{urltext}](tg://user?id={userid})", "{urltext}",
			user.FirstName+user.LastName, "{userid}", strconv.Itoa(user.Id))

		if user.Username != "" {
			name = html.EscapeString("@" + user.Username)
			text += fmt.Sprintf("\n - %s", name)
		}
	}
	_, err = u.EffectiveMessage.ReplyHTML(text)
	return err
}


func LoadAdmin(u *gotgbot.Updater) {
	defer log.Println("Loading module admin")
	u.Dispatcher.AddHandler(handlers.NewPrefixArgsCommand("promote", []rune{'/', '!'}, promote))
	u.Dispatcher.AddHandler(handlers.NewPrefixArgsCommand("demote", []rune{'/', '!'}, demote))
        u.Dispatcher.AddHandler(handlers.NewPrefixCommand("invitelink", []rune{'/', '!'}, invitelink))
	u.Dispatcher.AddHandler(handlers.NewPrefixCommand("adminlist", []rune{'/', '!'}, adminlist))
}
