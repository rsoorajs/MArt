package trans

import (
	"fmt"
	//gt "github.com/bas24/translategooglefree"
        "strings"
        "log"
        "github.com/PaulSonOfLars/gotgbot"
	"github.com/PaulSonOfLars/gotgbot/ext"
	"github.com/PaulSonOfLars/gotgbot/handlers"
	"github.com/PaulSonOfLars/gotgbot/parsemode"
        //"github.com/PaulSonOfLars/gotgbot/handlers/Filters"
)

func trans(bot ext.Bot, u *gotgbot.Update, args []string) error {
	chatId := u.EffectiveChat.Id
        text := u.EffectiveMessage.ReplyToMessage.Text
        from := u.EffectiveMessage.ReplyToMessage
        toText := strings.ToLower(args[0])
        replyId := u.EffectiveMessage.MessageId
        if args[0] == "" {
                toText = "en"
        }
        if from == nil {
                msg := bot.NewSendableMessage(chatId, "Reply to a message to Translate")
                msg.ReplyToMessageId = replyId
                _, m := msg.Send()
                return m
        }
        result, _ := Translate(text, "auto", toText)
        t := fmt.Sprintf("<u>Translated to <b>%s</b></u>\n\n<code>%s</code>", toText, result)
        msg := bot.NewSendableMessage(chatId, t)
        msg.ParseMode = parsemode.Html
        msg.ReplyToMessageId = replyId
        _, err := msg.Send()
        return err
}

func LoadTrans(u *gotgbot.Updater) {
	defer log.Println("Loading module Translator")
        u.Dispatcher.AddHandler(handlers.NewPrefixArgsCommand("tr", []rune{'/', '!'}, trans))
}

