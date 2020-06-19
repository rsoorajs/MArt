package filters

import (
	"fmt"
	"html"
	"log"
	"regexp"
	"strconv"
	"strings"

	tgmd2html "github.com/PaulSonOfLars/gotg_md2html"
	"github.com/PaulSonOfLars/gotgbot"
	"github.com/PaulSonOfLars/gotgbot/ext"
	"github.com/PaulSonOfLars/gotgbot/handlers"
	"github.com/PaulSonOfLars/gotgbot/handlers/Filters"
	"github.com/PaulSonOfLars/gotgbot/parsemode"
	"github.com/anandpskerala/Martha/bot/modules/sql"
	"github.com/anandpskerala/Martha/bot/modules/utils/chats"
	"github.com/anandpskerala/Martha/bot/modules/utils/error_handling"
	"github.com/anandpskerala/Martha/bot/modules/utils/extraction"
	"github.com/anandpskerala/Martha/bot/modules/utils/helpers"
)

func fget(bot ext.Bot, u *gotgbot.Update, noteName string, showNone bool, noFormat bool) error {
	chatId := u.EffectiveChat.Id
	note := sql.GetNote(strconv.Itoa(chatId), noteName)
	msg := u.EffectiveMessage

	replyId := msg.MessageId

	if note != nil {
		if msg.ReplyToMessage != nil {
			replyId = msg.ReplyToMessage.MessageId
		}

		if note.IsReply {
			msgId, _ := strconv.Atoi(note.Value)
			_, err := bot.ForwardMessage(chatId, chatId, msgId)
			if err != nil {
				_, err := msg.ReplyText("Looks like the original sender of this note has deleted " +
					"their message - sorry! I'll remove this note from " +
					"your saved notes.")
				sql.RmNote(strconv.Itoa(chatId), noteName)
				return err
			}
		} else {
			text := note.Value
			keyb := make([][]ext.InlineKeyboardButton, 0)
			buttons := sql.GetButtons(strconv.Itoa(chatId), noteName)
			parseMode := parsemode.Html
			btns := make([]tgmd2html.Button, len(buttons))
			for i, btn := range buttons {
				btns[i] = tgmd2html.Button{Name: btn.Name, Content: btn.Url, SameLine: btn.SameLine}
			}

			if noFormat {
				text = tgmd2html.Reverse(note.Value, btns)
				parseMode = ""
			} else {
				keyb = helpers.BuildKeyboard(buttons)
			}

			keyboard := &ext.InlineKeyboardMarkup{InlineKeyboard: &keyb}

			if note.Msgtype == sql.BUTTON_TEXT || note.Msgtype == sql.TEXT {
				msg := bot.NewSendableMessage(chatId, text)
				msg.ParseMode = parseMode
				msg.ReplyMarkup = keyboard
				msg.DisableWebPreview = true
				msg.ReplyToMessageId = replyId
				_, err := msg.Send()
				return err
			} else {

				var err error
				switch note.Msgtype {
				case sql.STICKER:
					msg := bot.NewSendableSticker(chatId)
					msg.ReplyToMessageId = replyId
					msg.FileId = note.File
					msg.ReplyMarkup = keyboard
					_, err = msg.Send()
					break
				case sql.DOCUMENT:
					msg := bot.NewSendableDocument(chatId, text)
					msg.ParseMode = parseMode
					msg.ReplyToMessageId = replyId
					msg.FileId = note.File
					msg.ReplyMarkup = keyboard
					_, err = msg.Send()
					break
				case sql.PHOTO:
					msg := bot.NewSendablePhoto(chatId, text)
					msg.ParseMode = parseMode
					msg.ReplyToMessageId = replyId
					msg.FileId = note.File
					msg.ReplyMarkup = keyboard
					_, err = msg.Send()
					break
				case sql.AUDIO:
					msg := bot.NewSendableAudio(chatId, text)
					msg.ParseMode = parseMode
					msg.ReplyToMessageId = replyId
					msg.FileId = note.File
					msg.ReplyMarkup = keyboard
					_, err = msg.Send()
					break
				case sql.VOICE:
					msg := bot.NewSendableVoice(chatId, text)
					msg.ParseMode = parseMode
					msg.ReplyToMessageId = replyId
					msg.FileId = note.File
					msg.ReplyMarkup = keyboard
					_, err = msg.Send()
					break
				case sql.VIDEO:
					msg := bot.NewSendableVideo(chatId, text)
					msg.ParseMode = parseMode
					msg.ReplyToMessageId = replyId
					msg.FileId = note.File
					msg.ReplyMarkup = keyboard
					_, err = msg.Send()
					break
				}

				if err != nil {
					if err.Error() == "Bad Request: Entity_mention_user_invalid" {
						_, _ = msg.ReplyText("Looks like you tried to mention someone I've never seen before. If you really " +
							"want to mention them, forward one of their messages to me, and I'll be able " +
							"to tag them!")
						return nil
					} else {
						_, _ = msg.ReplyText("This note could not be sent, as it is incorrectly formatted. Ask in " +
							"@Keralasbots if you can't figure out why!")
						return nil
					}
				}

			}
		}
	} else if showNone {
		_, err := msg.ReplyText("This note doesn't exist!")
		return err
	}
	return nil
}

func hashfGet(bot ext.Bot, u *gotgbot.Update) error {
	msg := u.EffectiveMessage
	chat := u.EffectiveChat
	filterList := sql.GetAllChatNotes(strconv.Itoa(chat.Id))
	if strings.HasPrefix(msg.Text, "/") == true {
		return gotgbot.ContinueGroups{}
	}

	toMatch := extraction.ExtractText(msg)
	matchText := strings.ToLower(toMatch)
	if toMatch == "" {
		return gotgbot.EndGroups{}
	}

	for _, trigger := range filterList {
		pattern := `( |^|[^\w])` + regexp.QuoteMeta(trigger.Name) + `( |$|[^\w])`
		re, err := regexp.Compile("(?i)" + pattern)
		error_handling.HandleErr(err)
		if re.MatchString(matchText) {
			triggers := trigger.Name
			return fget(bot, u, triggers, false, false)
		}
	}
	return nil
}

/*var customFiltertext handlers.FilterFunc = func(message *ext.Message) bool {
        return (Filters.Text(message)) && (Filters.Group(message))
}

var filterMessage = handlers.NewMessage(customFiltertext, hashfGet)
*/

func saveFilter(b ext.Bot, u *gotgbot.Update) error {
	chatId := strconv.Itoa(u.EffectiveChat.Id)
	msg := u.EffectiveMessage
	noteName, text, dataType, content, buttons := helpers.GetNoteType(msg)
	fnote := strings.ToLower(noteName)
	noteList := sql.GetAllChatNotes(chatId)

	if len(noteList) >= 5 {
		b.SendMessage(u.EffectiveChat.Id, "You have reached the limits")
		return nil
	}

	if !chats.RequireUserAdmin(u.EffectiveChat, msg, u.EffectiveUser.Id) {
		return gotgbot.ContinueGroups{}
	}

	if dataType == -1 {
		_, err := msg.ReplyText("Dude, there's no note!")
		return err
	}

	vchat := sql.GetChatConnect(chatId)

	if len(strings.TrimSpace(text)) == 0 {
		text = noteName
	}

	btns := make([]sql.Button, len(buttons))

	for i, btn := range buttons {
		btns[i] = sql.Button{ChatId: chatId, NoteName: noteName, Name: btn.Name, Url: btn.Content, SameLine: btn.SameLine}
	}

	if u.EffectiveChat.Type == "private" {
		if vchat.ChatId != "" {
			chattitle := vchat.ChatTitle
			go sql.AddNoteToDb(vchat.ChatId, fnote, text, dataType, btns, content)
			_, err := msg.ReplyHTMLf("Added <b>%v</b> to <b>%v</b>!", noteName, chattitle)
			return err
		}
	}
	go sql.AddNoteToDb(chatId, fnote, text, dataType, btns, content)
	_, err := msg.ReplyHTMLf("Added <b>%v</b>!", noteName)
	return err
}

func clearf(_ ext.Bot, u *gotgbot.Update, args []string) error {
	chatId := u.EffectiveChat.Id

	if !chats.RequireUserAdmin(u.EffectiveChat, u.EffectiveMessage, u.EffectiveUser.Id) {
		return gotgbot.ContinueGroups{}
	}
	vchat := sql.GetChatConnect(strconv.Itoa(chatId))

	if u.EffectiveChat.Type == "private" {
		if vchat.ChatId != "" {
			if len(args) >= 1 {
				noteName := strings.ToLower(args[0])
				if sql.RmNote(strconv.Itoa(chatId), noteName) {
					_, err := u.EffectiveMessage.ReplyHTMLf("Successfully removed <code>%v</code> from <b>%v</b>", noteName, vchat.ChatTitle)
					return err
				} else {
					_, err := u.EffectiveMessage.ReplyText("That's not a note in my database!")
					return err
				}
			}
		}
	}

	if len(args) >= 1 {
		noteName := strings.ToLower(args[0])

		if sql.RmNote(strconv.Itoa(chatId), noteName) {
			_, err := u.EffectiveMessage.ReplyHTMLf("Successfully removed <code>%v</code>", noteName)
			return err
		} else {
			_, err := u.EffectiveMessage.ReplyText("That's not a note in my database!")
			return err
		}
	}
	return nil
}

func listFilters(bot ext.Bot, u *gotgbot.Update) error {
	chatId := u.EffectiveChat.Id
	chat := u.EffectiveChat

	if u.EffectiveChat.Type == "private" {
		vchat := sql.GetChatConnect(strconv.Itoa(chatId))
		if vchat.ChatId != "" {
			chatId, _ = strconv.Atoi(vchat.ChatId)
			chat, _ = bot.GetChat(chatId)
		}
	}

	noteList := sql.GetAllChatNotes(strconv.Itoa(chatId))

	msg := fmt.Sprintf("List of Filters in <b>%v</b>\n", chat.Title)
	for _, note := range noteList {
		noteName := html.EscapeString(fmt.Sprintf(" - %v\n", note.Name))
		if len(msg)+len(noteName) > helpers.MaxMessageLength {
			_, err := u.EffectiveMessage.ReplyHTML(msg)
			msg = ""
			error_handling.HandleErr(err)
		}
		msg += noteName
	}

	if msg == fmt.Sprintf("List of Filters in %v\n", chat.Title) {
		_, err := u.EffectiveMessage.ReplyText("No filters in this chat!")
		return err
	} else if len(msg) != 0 {
		_, err := u.EffectiveMessage.ReplyHTML(msg)
		return err
	}
	return nil
}

var filterTextAndGroupFilter handlers.FilterFunc = func(message *ext.Message) bool {
	return (extraction.ExtractText(message) != "") && (Filters.Group(message))
}

func LoadFilters(u *gotgbot.Updater) {
	defer log.Println("Loading module filters")
	u.Dispatcher.AddHandlerToGroup(handlers.NewMessage(filterTextAndGroupFilter, hashfGet), 8)
	u.Dispatcher.AddHandler(handlers.NewPrefixCommand("filter", []rune{'/', '!'}, saveFilter))
	u.Dispatcher.AddHandler(handlers.NewPrefixArgsCommand("stop", []rune{'/', '!'}, clearf))
	u.Dispatcher.AddHandler(handlers.NewPrefixCommand("filters", []rune{'/', '!'}, listFilters))
}
