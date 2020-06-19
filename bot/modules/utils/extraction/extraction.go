package extraction

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/anandpskerala/Martha/bot/modules/users"
	"github.com/anandpskerala/Martha/bot/modules/utils/error_handling"
	"github.com/PaulSonOfLars/gotgbot/ext"
	"github.com/google/uuid"
)

func IdFromReply(m *ext.Message) (int, string) {
	prevMessage := m.ReplyToMessage
	if prevMessage == nil {
		return 0, ""
	}
	userId := prevMessage.From.Id
	res := strings.SplitN(m.Text, " ", 2)
	if len(res) < 2 {
		return userId, ""
	}
	return userId, res[1]
}

func ExtractUserAndText(m *ext.Message, args []string) (int, string) {
	prevMessage := m.ReplyToMessage
	splitText := strings.SplitN(m.Text, " ", 2)

	if len(splitText) < 2 {
		return IdFromReply(m)
	}

	textToParse := splitText[1]

	text := ""

	var userId int
	accepted := make(map[string]struct{})
	accepted["text_mention"] = struct{}{}

	entities := m.ParseEntityTypes(accepted)

	var ent *ext.ParsedMessageEntity
	var isId = false
	if len(entities) > 0 {
		ent = &entities[0]
	} else {
		ent = nil
	}

	if entities != nil && ent != nil && ent.Offset == (len(m.Text)-len(textToParse)) {
		ent = &entities[0]
		userId = ent.User.Id
		text = strconv.Itoa(int(m.Text[ent.Offset+ent.Length]))
	} else if len(args) >= 1 && args[0][0] == '@' {
		user := args[0]
		userId = users.GetUserId(user)
		if userId == 0 {
			_, err := m.ReplyText("I don't have that user in my db. You'll be able to interact with them if you reply to that person's message instead, or forward one of that user's messages.")
			error_handling.HandleErr(err)
			return 0, ""
		} else {
			res := strings.SplitN(m.Text, " ", 3)
			if len(res) >= 3 {
				text = res[2]
			}
		}
	} else if len(args) >= 1 {
		isId = true
		for _, arg := range args[0] {
			if unicode.IsDigit(arg) {
				continue
			} else {
				isId = false
				break
			}
		}
		if isId {
			userId, _ = strconv.Atoi(args[0])
			res := strings.SplitN(m.Text, " ", 3)
			if len(res) >= 3 {
				text = res[2]
			}
		}
	}
	if !isId && prevMessage != nil {
		_, parseErr := uuid.Parse(args[0])
		userId, text = IdFromReply(m)
		if parseErr == nil {
			return userId, text
		}
	} else if !isId {
		_, parseErr := uuid.Parse(args[0])
		if parseErr == nil {
			return userId, text
		}
	}

	_, err := m.Bot.GetChat(userId)
	if err != nil {

		_, err := m.ReplyText("I don't seem to have interacted with this user before - please forward a message from " +
			"them to give me control! (like a voodoo doll, I need a piece of them to be able " +
			"to execute certain commands...)")
		error_handling.HandleErr(err)
		return 0, ""
	}
	return userId, text
}

func ExtractUser(message *ext.Message, args []string) int {
	userId, _ := ExtractUserAndText(message, args)
	return userId
}

func ExtractText(message *ext.Message) string {
	if message.Text != "" {
		return message.Text
	} else if message.Caption != "" {
		return message.Caption
	} else if message.Sticker != nil {
		return message.Sticker.Emoji
	} else {
		return ""
	}
}
