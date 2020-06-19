package misc

import (
	"fmt"
	"html"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/anandpskerala/Martha/bot"
	//"github.com/anandpskerala/Martha/bot/modules/sql"
	"github.com/anandpskerala/Martha/bot/modules/utils/error_handling"
	"github.com/anandpskerala/Martha/bot/modules/utils/extraction"
	"github.com/anandpskerala/Martha/bot/modules/utils/helpers"
	"github.com/PaulSonOfLars/gotgbot"
	"github.com/PaulSonOfLars/gotgbot/ext"
	"github.com/PaulSonOfLars/gotgbot/handlers"
	"github.com/sirupsen/logrus"
	"github.com/tcnksm/go-httpstat"
)

func getId(bit ext.Bot, u *gotgbot.Update, args []string) error {
	userId := extraction.ExtractUser(u.EffectiveMessage, args)
	if userId != 0 {
		if u.EffectiveMessage.ReplyToMessage != nil && u.EffectiveMessage.ReplyToMessage.ForwardFrom != nil {
			user1 := u.EffectiveMessage.ReplyToMessage.From
			user2 := u.EffectiveMessage.ReplyToMessage.ForwardFrom
			_, err := u.EffectiveMessage.ReplyHTMLf("The ID of the original sender, %v, is <code>%v</code>.\n"+
				"The ID of the forwarder, %v, is <code>%v</code>.", html.EscapeString(user2.FirstName),
				user2.Id,
				html.EscapeString(user1.FirstName),
				user1.Id)
			return err
		} else {
			user, err := bit.GetChat(userId)
			error_handling.HandleErr(err)
			_, err = u.EffectiveMessage.ReplyHTMLf("%v's ID is <code>%v</code>", html.EscapeString(user.FirstName), user.Id)
		}
	} else {
		chat := u.EffectiveChat
		if chat.Type == "private" {
			_, err := u.EffectiveMessage.ReplyHTMLf("Your ID is <code>%v</code>", chat.Id)
			return err
		} else {
			_, err := u.EffectiveMessage.ReplyHTMLf("This group's ID is <code>%v</code>", chat.Id)
			return err
		}
	}
	return nil
}

func info(b ext.Bot, u *gotgbot.Update, args []string) error {
	msg := u.EffectiveMessage
	//chat := u.EffectiveChat
	userId := extraction.ExtractUser(msg, args)
	var user *ext.User

	if userId != 0 {
		userChat, _ := b.GetChat(userId)
		user = &ext.User{
			Id:        userChat.Id,
			FirstName: userChat.FirstName,
			LastName:  userChat.LastName,
		}

	} else if msg.ReplyToMessage == nil && len(args) <= 0 {
		user = msg.From
		userId = msg.From.Id

	} else if _, err := strconv.Atoi(args[0]); msg.ReplyToMessage == nil && (len(args) <= 0 || (len(args) >= 1 && strings.HasPrefix(args[0], "@") && err != nil && msg.ParseEntities()[0].Type != "TEXT_MENTION")) {
		_, err := msg.ReplyText("Nah, this mans doesn't exist.")
		return err
	} else {
		return nil
	}

	text := fmt.Sprintf("<b>User info</b>"+
		"\nID: <code>%v</code>"+
		"\nFirst Name: %v", userId, html.EscapeString(user.FirstName))

	if user.LastName != "" {
		text += fmt.Sprintf("\nLast Name: %v", user.LastName)
	}

	if user.Username != "" {
		text += fmt.Sprintf("\nUsername: @%v", user.Username)
	}

	text += fmt.Sprintf("\nPermanent user link: %v", helpers.MentionHtml(user.Id, user.FirstName+user.LastName))

	/*fed := sql.GetChatFed(strconv.Itoa(chat.Id))
	if fed != nil {
		fban := sql.GetFbanUser(fed.Id, strconv.Itoa(userId))
		if fban != nil {
			text += fmt.Sprintf("\n\nThis user is fedbanned in the current federation - "+
				"<code>%v</code>", fed.FedName)
		} else {
			text += "\n\nThis user is not fedbanned in the current federation."
		}
	}
*/

	if user.Id == bot.BotConfig.OwnerId {
		text += "\n\nThis is a strong man guys!!"
	} else {
		for _, id := range bot.BotConfig.SudoUsers {
			if strconv.Itoa(user.Id) == id {
				text += "\n\nThis person is one of my sudo users! " +
					"Powerful as my owner - so watch ot man."
			}
		}
	}
	_, err := u.EffectiveMessage.ReplyHTML(text)
	return err
}


func ping(_ ext.Bot, u *gotgbot.Update) error {
	req, err := http.NewRequest("GET", "https://google.com", nil)
	error_handling.HandleErr(err)

	var result httpstat.Result
	ctx := httpstat.WithHTTPStat(req.Context(), &result)
	req = req.WithContext(ctx)

	client := http.DefaultClient
	res, err := client.Do(req)
	error_handling.HandleErr(err)

	if _, err := io.Copy(ioutil.Discard, res.Body); err != nil {
		logrus.Println(err)
	}

	_ = res.Body.Close()

	text := fmt.Sprintf("Ping: <b>%d</b> ms", result.ServerProcessing/time.Millisecond)

	_, err = u.EffectiveMessage.ReplyHTML(text)
	return err
}

func LoadMisc(u *gotgbot.Updater) {
	defer log.Println("Loading module misc")
	u.Dispatcher.AddHandler(handlers.NewPrefixArgsCommand("id", []rune{'/', '!'}, getId))
	u.Dispatcher.AddHandler(handlers.NewPrefixArgsCommand("info", []rune{'/', '!'}, info))
	u.Dispatcher.AddHandler(handlers.NewPrefixCommand("ping", []rune{'/', '!'}, ping))
}
