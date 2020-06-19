package welcome

import (
	"fmt"
	"html"
	"log"
	"strconv"
	"strings"

	"github.com/anandpskerala/Martha/bot/modules/sql"
	"github.com/anandpskerala/Martha/bot/modules/utils/helpers"
	"github.com/PaulSonOfLars/gotgbot"
	"github.com/PaulSonOfLars/gotgbot/ext"
	"github.com/PaulSonOfLars/gotgbot/handlers"
	"github.com/PaulSonOfLars/gotgbot/handlers/Filters"
	"github.com/PaulSonOfLars/gotgbot/parsemode"
)

//var VALID_WELCOME_FORMATTERS = []string{"first", "last", "fullname", "username", "id", "count", "chatname", "mention"}

// EnumFuncMap map of welcome type to function to execute
var EnumFuncMap = map[int]func(ext.Bot, int, string) (*ext.Message, error){
	sql.TEXT:        ext.Bot.SendMessage,
	sql.BUTTON_TEXT: ext.Bot.SendMessage,
	sql.STICKER:     ext.Bot.SendStickerStr,
	sql.DOCUMENT:    ext.Bot.SendDocumentStr,
	sql.PHOTO:       ext.Bot.SendPhotoStr,
	sql.AUDIO:       ext.Bot.SendAudioStr,
	sql.VOICE:       ext.Bot.SendVoiceStr,
	sql.VIDEO:       ext.Bot.SendVideoStr,
}

func send(bot ext.Bot, u *gotgbot.Update, message string, keyboard *ext.InlineKeyboardMarkup, backupMessage string, reply bool) *ext.Message {
	msg := bot.NewSendableMessage(u.EffectiveChat.Id, message)
	msg.ParseMode = parsemode.Html
	if reply {
		msg.ReplyToMessageId = u.EffectiveMessage.MessageId
	}
	msg.ReplyMarkup = keyboard
	m, err := msg.Send()
	if err != nil {
		m, _ = u.EffectiveMessage.ReplyText(backupMessage + "Note: The current message was invalid due to some issues.")
	}
	return m
}

func newMember(bot ext.Bot, u *gotgbot.Update) error {
	chat := u.EffectiveChat
	newMembers := u.EffectiveMessage.NewChatMembers
	welcPrefs := sql.GetWelcomePrefs(strconv.Itoa(chat.Id))
	var firstName = ""
	var fullName = ""
	var username = ""
	var res = ""
	var keyb = make([][]ext.InlineKeyboardButton, 0)

	if welcPrefs.ShouldWelcome {
		for _, mem := range newMembers {
			if mem.Id == bot.Id {
				continue
			}

			if welcPrefs.WelcomeType != sql.TEXT && welcPrefs.WelcomeType != sql.BUTTON_TEXT {
				_, err := EnumFuncMap[welcPrefs.WelcomeType](bot, chat.Id, welcPrefs.CustomWelcome)
				if err != nil {
					return err
				}
			}
			firstName = mem.FirstName
			if len(mem.FirstName) <= 0 {
				firstName = "PersonWithNoName"
			}

			if welcPrefs.CustomWelcome != "" {
				if mem.LastName != "" {
					fullName = firstName + " " + mem.LastName
				} else {
					fullName = firstName
				}
				count, _ := chat.GetMembersCount()
				mention := helpers.MentionHtml(mem.Id, firstName)

				if mem.Username != "" {
					username = "@" + html.EscapeString(mem.Username)
				} else {
					username = mention
				}

				r := strings.NewReplacer(
					"{first}", html.EscapeString(firstName),
					"{last}", html.EscapeString(mem.LastName),
					"{fullname}", html.EscapeString(fullName),
					"{username}", username,
					"{mention}", mention,
					"{count}", strconv.Itoa(count),
					"{chatname}", html.EscapeString(chat.Title),
					"{id}", strconv.Itoa(mem.Id),
					"{rules}", "",
				)
				res = r.Replace(welcPrefs.CustomWelcome)
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
				}
				keyb = helpers.BuildWelcomeKeyboard(buttons)
			} else {
				r := strings.NewReplacer("{first}", firstName)
				res = r.Replace(sql.DefaultWelcome)
			}

			if welcPrefs.ShouldMute {
				if !sql.IsUserHuman(strconv.Itoa(mem.Id), strconv.Itoa(chat.Id)) {
					if !sql.HasUserClickedButton(strconv.Itoa(mem.Id), strconv.Itoa(chat.Id)) {
						_, _ = bot.RestrictChatMember(chat.Id, mem.Id)
					}
				}
				kb := make([]ext.InlineKeyboardButton, 1)
				kb[0] = ext.InlineKeyboardButton{Text: "Click here to prove you're human", CallbackData: "unmute"}
				keyb = append(keyb, kb)
			}

			keyboard := &ext.InlineKeyboardMarkup{InlineKeyboard: &keyb}
			r := strings.NewReplacer("{first}", firstName)
			sent := send(bot, u, res, keyboard, r.Replace(sql.DefaultWelcome), !welcPrefs.DelJoined)

			if welcPrefs.CleanWelcome != 0 {
				_, _ = bot.DeleteMessage(chat.Id, welcPrefs.CleanWelcome)
				sql.SetCleanWelcome(strconv.Itoa(chat.Id), sent.MessageId)
			}

			if welcPrefs.DelJoined {
				_, _ = u.EffectiveMessage.Delete()
			}
		}
	}
	return nil
}

func unmuteCallback(bot ext.Bot, u *gotgbot.Update) error {
	query := u.CallbackQuery
	user := u.EffectiveUser
	chat := u.EffectiveChat

	if !sql.IsUserHuman(strconv.Itoa(user.Id), strconv.Itoa(chat.Id)) {
		if !sql.HasUserClickedButton(strconv.Itoa(user.Id), strconv.Itoa(chat.Id)) {
			_, err := bot.UnRestrictChatMember(chat.Id, user.Id)
			if err != nil {
				return err
			}
			go sql.UserClickedButton(strconv.Itoa(user.Id), strconv.Itoa(chat.Id))
			_, _ = bot.AnswerCallbackQueryText(query.Id, "You've proved that you are a human! "+
				"You can now talk in this group.", false)
			return nil
		}
	}

	_, _ = bot.AnswerCallbackQueryText(query.Id, "This action is invalid for you.", false)
	return gotgbot.EndGroups{}
}

// LoadWelcome load welcome module
func LoadWelcome(u *gotgbot.Updater) {
	defer log.Println("Loading module welcome")
	u.Dispatcher.AddHandler(handlers.NewMessage(Filters.NewChatMembers(), newMember))
	u.Dispatcher.AddHandler(handlers.NewCallback("unmute", unmuteCallback))
	u.Dispatcher.AddHandler(handlers.NewPrefixArgsCommand("welcome", []rune{'!', '/'}, welcome))
	u.Dispatcher.AddHandler(handlers.NewPrefixCommand("setwelcome", []rune{'!', '/'}, setWelcome))
	u.Dispatcher.AddHandler(handlers.NewPrefixCommand("resetwelcome", []rune{'!', '/'}, resetWelcome))
	u.Dispatcher.AddHandler(handlers.NewPrefixArgsCommand("cleanwelcome", []rune{'!', '/'}, cleanWelcome))
	u.Dispatcher.AddHandler(handlers.NewPrefixArgsCommand("deljoined", []rune{'!', '/'}, delJoined))
	u.Dispatcher.AddHandler(handlers.NewPrefixArgsCommand("welcomemute", []rune{'!', '/'}, welcomeMute))
}
