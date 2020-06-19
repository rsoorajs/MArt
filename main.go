package main

import (
	"log"
        "fmt"
	"os"
        //"unicode"
	"strconv"
        "regexp"
        "html"
        "strings"
	
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	
	"github.com/PaulSonOfLars/gotgbot"
        "github.com/PaulSonOfLars/gotgbot/ext"
        "github.com/PaulSonOfLars/gotgbot/parsemode"
	//"github.com/PaulSonOfLars/gotgbot/ext"
	"github.com/PaulSonOfLars/gotgbot/handlers"
        "github.com/anandpskerala/Martha/bot"
	"github.com/anandpskerala/Martha/bot/modules/utils/caching"
	"github.com/anandpskerala/Martha/bot/modules/utils/error_handling"
        "github.com/anandpskerala/Martha/bot/modules/pin"
        "github.com/anandpskerala/Martha/bot/modules/admin"
        "github.com/anandpskerala/Martha/bot/modules/misc"
        "github.com/anandpskerala/Martha/bot/modules/users"
        "github.com/anandpskerala/Martha/bot/modules/sql"
        "github.com/anandpskerala/Martha/bot/modules/rules"
        "github.com/anandpskerala/Martha/bot/modules/welcome"
        "github.com/anandpskerala/Martha/bot/modules/utils/helpers"
        "github.com/anandpskerala/Martha/bot/modules/ban"
        "github.com/anandpskerala/Martha/bot/modules/feds"
        "github.com/anandpskerala/Martha/bot/modules/notes"
	//tg_md2html "github.com/PaulSonOfLars/gotg_md2html"
        "github.com/anandpskerala/Martha/bot/modules/warns"
        "github.com/anandpskerala/Martha/bot/modules/blacklist"
        "github.com/anandpskerala/Martha/bot/modules/purge"
        "github.com/anandpskerala/Martha/bot/modules/cust_filters"
        "github.com/anandpskerala/Martha/bot/modules/report"
        "github.com/anandpskerala/Martha/bot/modules/mutes"
        "github.com/anandpskerala/Martha/bot/modules/trans"
	"github.com/anandpskerala/Martha/bot/modules/connection"
)

func main() {
	//Logger
	cfg := zap.NewProductionEncoderConfig()
	cfg.EncodeLevel = zapcore.CapitalLevelEncoder
	cfg.EncodeTime = zapcore.RFC3339TimeEncoder

	logger := zap.New(zapcore.NewCore(zapcore.NewConsoleEncoder(cfg), os.Stdout, zap.InfoLevel))
	defer logger.Sync() // flushes buffer, if any
	l := logger.Sugar()

	// Create updater instance
	u, err := gotgbot.NewUpdater(logger, bot.BotConfig.ApiKey)
	error_handling.FatalError(err)
	l.Info("Starting Martha...")

	// Add start handler
	u.Dispatcher.AddHandler(handlers.NewArgsCommand("start", start))
	u.Dispatcher.AddHandler(handlers.NewCommand("donate", donate))

	// Create database tables if not already existing
	sql.EnsureBotInDb(u)

	// Prepare Caching Service
	caching.InitCache()

        //Module handler
        pin.LoadPin(u)
        misc.LoadMisc(u)
        users.LoadUsers(u)
        admin.LoadAdmin(u)
        rules.LoadRules(u)
        welcome.LoadWelcome(u)
        bans.LoadBans(u)
        feds.LoadFeds(u)
        notes.LoadNotes(u)
        warns.LoadWarns(u)
        blacklist.LoadBlacklist(u)
        purge.LoadPurge(u)
        filters.LoadFilters(u)
        report.LoadReport(u)
        mute.LoadMute(u)
        trans.LoadTrans(u)
	connection.LoadConnect(u)
        LoadHelp(u)

	if bot.BotConfig.DropUpdate == "True" {
		log.Println("[Info][Core] Using Clean Long Polling")
		err = u.StartCleanPolling()
		error_handling.HandleErr(err)
	} else {
		log.Println("[Info][Core] Using Long Polling")
		err = u.StartPolling()
		error_handling.HandleErr(err)
	}

	u.Idle()
}

func isNumeric(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}


func start(b ext.Bot, u *gotgbot.Update, args []string) error {
	msg := u.EffectiveMessage

	if u.EffectiveChat.Type == "private" {        
		if len(args) != 0 {
                        s := isNumeric(args[0])
                        //r := rune(s)
                        if s == true {
			        if _, err := strconv.Atoi(args[0][2:]); err == nil {
				        chatRules := sql.GetChatRules(args[0])
				        if chatRules != nil {
					        _, err := msg.ReplyHTML(chatRules.Rules)
					        return err
				        }
				        _, err := msg.ReplyText("The admins of the group haven't set any rules for this chat yet. This probably doesn't " +
					        "mean it's lawless though!")
				        log.Println(args[0])
				        return err
			        }
		        }
                        if strings.ToLower(args[0]) == "help" {
                                if u.EffectiveChat.Type == "private" {
	                                msg := b.NewSendableMessage(u.EffectiveChat.Id, "Hey there! I'm <b>Martha</b>, a group management bot written in Go Language.\n"+
	       	                                "Younger sister of @Miss_Miabot. I have a lot of features like Filters, Notes, Warns, Mutes, etc\n\n"+
		                                "Some Handy commands are :\n\n"+
		                                " - /start: you already know what this does\n\n"+
		                                " - /help: For info on how to use me\n\n"+
		                                " - /donate: To support my owner\n\n\n"+
		                                "If you have any bugs reports, queries or suggestions you can head over to @keralasbots.\n\n"+
		                                "All commands can be used with the following: / or !")
	                                msg.ParseMode = parsemode.Html
	                                msg.ReplyToMessageId = u.EffectiveMessage.MessageId
	                                msg.ReplyMarkup = &markup
	                                _, err := msg.Send()
                                        if err != nil {
		                                msg.ReplyToMessageId = 0
		                                _, err = msg.Send()
	                                }
	                                return err
                                }
                        }
	        }
        }

	if u.EffectiveChat.Type != "private" {
                _, err := msg.ReplyText("Hi Friends :)")
                return err
        }

	button := sql.WelcomeButton{
		Id:       0,
		ChatId:   strconv.Itoa(u.EffectiveChat.Id),
		Name:     "Add Me to Group",
		Url:      fmt.Sprintf("t.me/%v?startgroup=True", b.UserName),
		SameLine: false,
	}
	keyb := helpers.BuildWelcomeKeyboard([]sql.WelcomeButton{button})
	keyboard := ext.InlineKeyboardMarkup{InlineKeyboard: &keyb}

	bot := b.NewSendableMessage(u.EffectiveChat.Id, "Hi there! I'm  Martha, younger sister of @Miss_Miabot. I am a group management bot, written in Go Language."+
		"\nIf you have an queries or doubts regarding me, you can ask it in @KeralasBots.\nType /help to get details to use me. If you are interested in donating hit /donate. It will be great if you donate ")

	bot.ParseMode = parsemode.Html
	bot.ReplyToMessageId = u.EffectiveMessage.MessageId
	bot.ReplyMarkup = &keyboard
	_, err := bot.Send()
	if err != nil {
		bot.ReplyToMessageId = 0
		_, err = bot.Send()
	}
	return err
}

var markup ext.InlineKeyboardMarkup
var markdownHelpText string

func initMarkdownHelp() {
	markdownHelpText = "You can use markdown to make your messages more expressive. This is the markdown currently " +
		"supported:\n\n" +
		"<code>`code words`</code>: backticks allow you to wrap your words in monospace fonts.\n" +
		"<code>*bold*</code>: wrapping text with '*' will produce bold text\n" +
		"<code>_italics_</code>: wrapping text with '_' will produce italic text\n" +
		"<code>[hyperlink](example.com)</code>: this will create a link - the message will just show " +
		"<code>hyperlink</code>, and tapping on it will open the page at <code>example.com</code>\n\n" +
		"<code>[buttontext](buttonurl:example.com)</code>: this is a special enhancement to allow users to have " +
		"telegram buttons in their markdown. <code>buttontext</code> will be what is displayed on the button, and " +
		"<code>example.com</code> will be the url which is opened.\n\n" +
		"If you want multiple buttons on the same line, use :same, as such:\n" +
		"<code>[one](buttonurl://github.com)</code>\n" +
		"<code>[two](buttonurl://google.com:same)</code>\n" +
		"This will create two buttons on a single line, instead of one button per line.\n\n" +
		"Keep in mind that your message MUST contain some text other than just a button!"

}

func initHelpButtons() {
	helpButtons := [][]ext.InlineKeyboardButton{make([]ext.InlineKeyboardButton, 2), make([]ext.InlineKeyboardButton, 2),
		make([]ext.InlineKeyboardButton, 2), make([]ext.InlineKeyboardButton, 2), make([]ext.InlineKeyboardButton, 2), make([]ext.InlineKeyboardButton, 2)}

	// First column
	helpButtons[0][0] = ext.InlineKeyboardButton{
		Text:         "Admins",
		CallbackData: fmt.Sprintf("help(%v)", "admin"),
	}
	helpButtons[1][0] = ext.InlineKeyboardButton{
		Text:         "Bans",
		CallbackData: fmt.Sprintf("help(%v)", "bans"),
	}
	helpButtons[2][0] = ext.InlineKeyboardButton{
		Text:         "Blacklist",
		CallbackData: fmt.Sprintf("help(%v)", "blacklist"),
	}
	helpButtons[3][0] = ext.InlineKeyboardButton{
		Text:         "Deleting",
		CallbackData: fmt.Sprintf("help(%v)", "deleting"),
	}
	helpButtons[4][0] = ext.InlineKeyboardButton{
		Text:         "Federations",
		CallbackData: fmt.Sprintf("help(%v)", "feds"),
	}
	helpButtons[5][0] = ext.InlineKeyboardButton{
		Text:         "Filters",
		CallbackData: fmt.Sprintf("help(%v)", "filter"),
	}
        //Second column

	helpButtons[0][1] = ext.InlineKeyboardButton{
		Text:         "Greetings",
		CallbackData: fmt.Sprintf("help(%v)", "welcome"),
	}
	helpButtons[1][1] = ext.InlineKeyboardButton{
		Text:         "Markdown",
		CallbackData: fmt.Sprintf("help(%v)", "markdown"),
	}
	helpButtons[2][1] = ext.InlineKeyboardButton{
		Text:         "Misc",
		CallbackData: fmt.Sprintf("help(%v)", "misc"),
	}
	helpButtons[3][1] = ext.InlineKeyboardButton{
		Text:         "Mute",
		CallbackData: fmt.Sprintf("help(%v)", "muting"),
	}
	helpButtons[4][1] = ext.InlineKeyboardButton{
		Text:         "Notes",
		CallbackData: fmt.Sprintf("help(%v)", "notes"),
	}
	helpButtons[5][1] = ext.InlineKeyboardButton{
		Text:         "Warnings",
		CallbackData: fmt.Sprintf("help(%v)", "warns"),
	}

	markup = ext.InlineKeyboardMarkup{InlineKeyboard: &helpButtons}

}

func help(b ext.Bot, u *gotgbot.Update) error {
        if u.EffectiveChat.Type == "private" {
	        msg := b.NewSendableMessage(u.EffectiveChat.Id, "Hey there! I'm <b>Martha</b>, Younger sister of @Miss_Miabot. I am a group management bot written in Go Language.\n"+
	       	        "I have a lot of features like Filters, Notes, Warns, Mutes, etc.\n\n"+
		        "Some Handy commands are :\n\n"+
		        " - /start: you already know what this does\n\n"+
		        " - /help: For info on how to use me\n\n"+
		        " - /donate: To support my owner\n\n\n"+
		        "If you have any bugs reports, queries or suggestions you can head over to @keralasbots.\n\n"+
		        "All commands can be used with the following: / or !")
	        msg.ParseMode = parsemode.Html
	        msg.ReplyToMessageId = u.EffectiveMessage.MessageId
	        msg.ReplyMarkup = &markup
	        _, err := msg.Send()
                if err != nil {
		        msg.ReplyToMessageId = 0
		        _, err = msg.Send()
	        }
	        return err
        }
        msg := b.NewSendableMessage(u.EffectiveChat.Id, "Contact me in PM to get help message.")
	button := sql.WelcomeButton{
	        Id:       0,
		ChatId:   strconv.Itoa(u.EffectiveChat.Id),
		Name:     "Help",
		Url:      fmt.Sprintf("t.me/%v?start=help", b.UserName),
		SameLine: false,
	}
	keyb := helpers.BuildWelcomeKeyboard([]sql.WelcomeButton{button})
	keyboard := ext.InlineKeyboardMarkup{InlineKeyboard: &keyb}
	msg.ReplyMarkup = &keyboard
	_, err := msg.Send()
	return err
}

func markdownHelp(_ ext.Bot, u *gotgbot.Update) error {
	chat := u.EffectiveChat
	if chat.Type != "private" {
		_, err := u.EffectiveMessage.ReplyText("This command is meant to be used in PM!")
		return err
	}

	_, err := u.EffectiveMessage.ReplyHTML(markdownHelpText)
	return err
}

func buttonHandler(b ext.Bot, u *gotgbot.Update) error {
	query := u.CallbackQuery
	pattern, _ := regexp.Compile(`help\((.+?)\)`)

	if pattern.MatchString(query.Data) {
		module := pattern.FindStringSubmatch(query.Data)[1]
		chat := u.EffectiveChat
		msg := b.NewSendableEditMessageText(chat.Id, u.EffectiveMessage.MessageId, "placeholder")
		msg.ParseMode = parsemode.Html
		backButton := [][]ext.InlineKeyboardButton{{ext.InlineKeyboardButton{
			Text:         "🔙 Back",
			CallbackData: "help(back)",
		}}}
		backKeyboard := ext.InlineKeyboardMarkup{InlineKeyboard: &backButton}
		msg.ReplyMarkup = &backKeyboard

		switch module {
		case "admin":
			msg.Text = "Here is the help for the <b>Admin</b> module:\n\n" +
				"- /adminlist: list of admins in the chat\n\n" +
				"<b>Admin only:</b>\n" +
				html.EscapeString("- /pin: Pins the message.\n"+
					"- /unpin: unpins the currently pinned message\n"+
					"- /invitelink: gets invitelink\n"+
					"- /promote: promotes the user replied to\n"+
					"- /demote: demotes the user replied to\n")
			break
		case "bans":
			msg.Text = "Here is the help for the <b>Bans</b> module:\n\n" +
				" - /kickme: kicks the user who uses the command\n\n" +
				"<b>Admin only</b>:\n" +
				html.EscapeString(" - /ban <userhandle>: bans a user. (via handle, or reply)\n"+
					" - /tban <userhandle> x(m/h/d): bans a user for x time. (via handle, or reply). m = minutes, h = hours,"+
					" d = days.\n"+
					"- /unban <userhandle>: unbans a user. (via handle, or reply)"+
					" - /kick <userhandle>: kicks a user, (via handle, or reply)")

			break
		case "blacklist":
			msg.Text = "Here is the help for the <b>Word Blacklists</b> module:\n\n" +
				"Blacklists are used to stop certain triggers from being said in a group. Any time the trigger is " +
				"mentioned, the message will immediately be deleted. A good combo is sometimes to pair this up with " +
				"warn filters!\n\n" +
				"<b>NOTE:</b> blacklists do not affect group admins.\n\n" +
				" - /blacklist: View the current blacklisted words.\n\n" +
				"<b>Admin only:</b>\n" +
				html.EscapeString("- /addblacklist <triggers>: Add a trigger to the blacklist. Each line is "+
					"considered one trigger, so using different lines will allow you to add multiple triggers.\n"+
					"- /unblacklist <triggers>: Remove triggers from the blacklist. Same newline logic applies here, "+
					"so you can remove multiple triggers at once.\n"+
					" - /rmblacklist <triggers>: Same as above.")
			break
		case "deleting":
			msg.Text = "Here is the help for the <b>Purges</b> module:\n\n" +
				"<b>Admin only:</b>\n" +
				" - /del: deletes the message you replied to\n" +
				" - /purge: deletes all messages between this and the replied to message.\n"
			break
		case "feds":
                        msg.Text = "Here is the help for the <b>Federation</b> module:\n\n" +
                                " - /fedinfo: Gets information about the feds\n" +
                                " - /newfed [fedname]: Creates new fed.\n" +
                                " - /joinfed : Joins a fed.\n" +
                                " - /deletefed: To delete a fed.\n"
			break
		case "misc":
                        msg.Text = "Hers is the help for the <b>Misc</b> module:\n\n" +
                                " - /id: Gets your id\n" +
                                " - /ping: Gets a ping message\n" +
                                " - /info: Gets the info of the user\n"
			break
		case "muting":
			msg.Text = "Here is the help for the <b>Muting</b> module:\n\n" +
				"<b>Admin only:</b>\n" +
				html.EscapeString("- /mute <userhandle>: silences a user. Can also be used as a reply, muting the "+
					"replied to user.\n"+
					"- /tmute <userhandle> x(m/h/d): mutes a user for x time. (via handle, or reply). m = minutes, h = "+
					"hours, d = days.\n"+
					"- /unmute <userhandle>: unmutes a user. Can also be used as a reply, muting the replied to user.")
			break
                case "filter":
                        msg.Text = "Here is the help for <b>Filter</b> Module:\n\n <b>Available commands are:</b>\n\n" +
                                html.EscapeString(" - /filter <word> <sentence>: Every time someone says word, the bot will reply with sentence. For multiple word filters, quote the first word.\n" +
                                " - /filters: List all filters active in the current chat.\n" +
                                " - /stop <word>: Stop the bot replying to word.\n")
                        break
                case "markdown":
                        msg.Text = "You can use markdown to make your messages more expressive. This is the markdown currently " +
		                "Supported Formats:\n\n" +
		                "<code>`code words`</code>: backticks allow you to wrap your words in monospace fonts.\n" +
		                "<code>*bold*</code>: wrapping text with '*' will produce bold text\n" +
		                "<code>_italics_</code>: wrapping text with '_' will produce italic text\n" +
		                "<code>[hyperlink](example.com)</code>: this will create a link - the message will just show " +
		                "<code>hyperlink</code>, and tapping on it will open the page at <code>example.com</code>\n\n" +
		                "<code>[buttontext](buttonurl:example.com)</code>: this is a special enhancement to allow users to have " +
		                "telegram buttons in their markdown. <code>buttontext</code> will be what is displayed on the button, and " +
		                "<code>example.com</code> will be the url which is opened.\n\n" +
		                "If you want multiple buttons on the same line, use :same, as such:\n" +
		                "<code>[one](buttonurl://github.com)</code>\n" +
		                "<code>[two](buttonurl://google.com:same)</code>\n" +
		                "This will create two buttons on a single line, instead of one button per line.\n\n" +
		                "Keep in mind that your message MUST contain some text other than just a button!"
                        break
		case "notes":
			msg.Text = "Here is the help for the <b>Notes</b> module:\n\n" +
				html.EscapeString("- /get <notename>: get the note with this notename\n"+
					"- #<notename>: same as /get\n"+
					"- /notes or /saved: list all saved notes in this chat\n\n"+
					"If you would like to retrieve the contents of a note without any formatting, use /get"+
					" <notename> noformat. This can be useful when updating a current note.\n\n") +
				"<b>Admin only:</b>\n" +
				html.EscapeString(" - /save <notename> <notedata>: saves notedata as a note with name notename\n"+
					"A button can be added to a note by using standard markdown link syntax - the link should just "+
					"be prepended with a buttonurl: section, as such: [somelink](buttonurl:example.com). Check "+
					"/markdownhelp for more info.\n"+
					" - /save <notename>: save the replied-to message as a note with name notename\n"+
					" - /clear <notename>: clear note with this name")
			break
		case "welcome":
                        msg.Text = " Here is the help for the <b>Greetings</b> module:\n\n" +
                                html.EscapeString(" - /welcome: To get the current welcome message\n" +
                                        " - /setwelcome: To set a new welcome message \n" +
                                        " - /resetwelcome: To reset the welcome message \n" +
                                        " - /cleanwelcome <on/off>: On new member, try to delete the previous welcome message to avoid spamming the chat.\n" +
                                        " - /welcomemute <on/off/yes/no>: all users that join, get muted; a button gets added to the welcome message for them to unmute themselves. This proves they aren't a bot!")
			break
		case "warns":
			msg.Text = "Here is the help for the <b>Warnings</b> module:\n\n" +
				html.EscapeString(" - /warns <userhandle>: get a user's number, and reason, of warnings.\n"+
					" - /warnlist: list of all current warning filters\n\n") +
				"<b>Admin only:</b>\n" +
				html.EscapeString("- /warn <userhandle>: warn a user. After the warn limit, the user will be banned from the group. "+
					"Can also be used as a reply.\n"+
					" - /resetwarn <userhandle>: reset the warnings for a user. Can also be used as a reply.\n"+
					" - /addwarn <keyword> <reply message>: set a warning filter on a certain keyword. If you want your "+
					"keyword to be a sentence, encompass it with quotes, as such: /addwarn \"very angry\" "+
					"This is an angry user.\n"+
					"- /nowarn <keyword>: stop a warning filter\n"+
					"- /warnlimit <num>: set the warning limit\n"+
					" - /strongwarn <on/yes/off/no>: If set to on, exceeding the warn limit will result in a ban. "+
					"Else, will just kick.\n")
			break
		case "back":
			msg.Text = "Hey there! I'm <b>Martha</b>, Younger sister of @Miss_Miabot. I am a group management bot written in Go Language.\n" +
				"I have a lot of features like Filters, Notes, Warns, Mutes, etc.\n\n"+
		                "Some Handy commands are :\n\n"+
		                " - /start: you already know what this does\n\n"+
		                " - /help: For info on how to use me\n\n"+
		                " - /donate: To support my owner\n\n\n"+
		                "If you have any bugs reports, queries or suggestions you can head over to @keralasbots.\n\n"+
		                "All commands can be used with the following: / or !"
			msg.ReplyMarkup = &markup
			break
		}

		_, err := msg.Send()
		error_handling.HandleErr(err)
		_, err = b.AnswerCallbackQuery(query.Id)
		return err
	}
	return nil
}

func LoadHelp(u *gotgbot.Updater) {
	defer log.Println("Loading module help")
	initHelpButtons()
	initMarkdownHelp()
	u.Dispatcher.AddHandler(handlers.NewPrefixCommand("help", []rune{'/', '!'}, help))
	u.Dispatcher.AddHandler(handlers.NewCallback("help", buttonHandler))
	u.Dispatcher.AddHandler(handlers.NewPrefixCommand("markdownhelp", []rune{'/', '!'}, markdownHelp))
}

func donate(b ext.Bot, u *gotgbot.Update) error {
	_, err := u.EffectiveMessage.ReplyHTML("It took a lot of work for my creator to get me to where I am now - so if you have some money to spare, and want to show your support; Donate!" +
		"After all, server fees don't pay themselves - so every little helps! All donation money goes straight to funding the VPS." +
		"\n\n You can donate through UPI Method.\n My owner's UPI Address : <code>anandps002@oksbi</code>")
	return err
}

