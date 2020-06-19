package report

import (
	"fmt"
	"log"


	"github.com/anandpskerala/Martha/bot/modules/utils/chats"
	"github.com/anandpskerala/Martha/bot/modules/utils/error_handling"
        "github.com/PaulSonOfLars/gotgbot"
	"github.com/PaulSonOfLars/gotgbot/ext"
	"github.com/PaulSonOfLars/gotgbot/handlers"
        "github.com/PaulSonOfLars/gotgbot/parsemode"
)



func report(b ext.Bot, u *gotgbot.Update) error {
	chat := u.EffectiveChat
        fromUser := u.EffectiveMessage.ReplyToMessage.From
        fromId := u.EffectiveMessage.ReplyToMessage.MessageId
        byUser := u.EffectiveMessage.From
	if u.EffectiveChat.Type == "private" {
		_, err := u.EffectiveMessage.ReplyText("This command is meant to be used in a group!")
		return err
	}
        if chats.IsUserAdmin(chat, u.EffectiveUser.Id) {
		return gotgbot.ContinueGroups{}
	}
	

	admins, err := u.EffectiveChat.GetAdministrators()
	error_handling.HandleErr(err)

        //if u.EffectiveChat.Title != "" {
        chattitle := u.EffectiveChat.Title
	//}
        cuser := ""
       // usid := strconv.Atoi("")
        if u.EffectiveChat.Username != "" {
                cuser += chat.Username
        } else {
                cuser += fmt.Sprintf("c/%d", chat.Id)
        }
	text := fmt.Sprintf("*%v:*\n\n*Reported User : *[%s](tg://user?id=%d)\n*Reported By : *[%s](tg://user?id=%d)\n*Link : *[Click Here](https://t.me/%s/%d)", chattitle, fromUser.FirstName, fromUser.Id, byUser.FirstName, byUser.Id, cuser, fromId)

	for _, admin := range admins {
		user := admin.User
                usid := user.Id
                if user.IsBot == true {
                        continue;
                }
		
		_, err = u.EffectiveMessage.ReplyMarkdownf("Reported [%s](tg://user?id=%d) to Admins", fromUser.FirstName, fromUser.Id)
                msg := b.NewSendableMessage(usid, text)
                msg.ParseMode = parsemode.Markdown
                _, err = msg.Send()
                if err.Error() == "Forbidden: bot can't initiate conversation with a user" {
                        continue;
                }
	        return err
        }
        return nil
	
}

func LoadReport(u *gotgbot.Updater) {
	defer log.Println("Loading module Report")
	u.Dispatcher.AddHandler(handlers.NewPrefixCommand("report", []rune{'/', '!'}, report))
	u.Dispatcher.AddHandler(handlers.NewRegex(`(?i)@admin*`, report))

}
