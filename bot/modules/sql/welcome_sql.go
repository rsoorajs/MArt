package sql

const DefaultWelcome = "Hey {first}, how are you?"

type Welcome struct {
	ChatId        string `gorm:"primary_key"`
	CustomWelcome string
	ShouldWelcome bool `gorm:"default:true"`
	ShouldMute    bool `gorm:"default:true"`
	DelJoined     bool `gorm:"default:false"`
	CleanWelcome  int  `gorm:"default:0"`
	WelcomeType   int  `gorm:"default:0"`
	MuteTime      int  `gorm:"default:0"`
}

type WelcomeButton struct {
	Id       uint   `gorm:"primary_key;AUTO_INCREMENT"`
	ChatId   string `gorm:"primary_key"`
	Name     string `gorm:"not null"`
	Url      string `gorm:"not null"`
	SameLine bool   `gorm:"default:false"`
}

type MutedUser struct {
	UserId        string `gorm:"primary_key"`
	ChatId        string `gorm:"primary_key"`
	ButtonClicked bool   `gorm:"default:false"`
}

// GetWelcomePrefs Return the preferences for welcoming users
func GetWelcomePrefs(chatID string) *Welcome {
	welc := &Welcome{ChatId: chatID}

	if SESSION.First(welc).RowsAffected == 0 {
		return &Welcome{
			ChatId:        chatID,
			ShouldWelcome: true,
			ShouldMute:    false,
			CleanWelcome:  0,
			DelJoined:     false,
			CustomWelcome: DefaultWelcome,
			WelcomeType:   TEXT,
			MuteTime:      0,
		}
	}
	return welc
}

// GetWelcomeButtons Get the buttons for the welcome message
func GetWelcomeButtons(chatID string) []WelcomeButton {
	var buttons []WelcomeButton
	SESSION.Where("chat_id = ?", chatID).Find(&buttons)
	return buttons
}

// SetCleanWelcome Set whether to clean old welcome messages or not
func SetCleanWelcome(chatID string, cw int) {
	w := &Welcome{ChatId: chatID}
	tx := SESSION.Begin()
	tx.FirstOrCreate(w)
	w.CleanWelcome = cw
	tx.Save(w)
	tx.Commit()
}

// GetCleanWelcome Get whether to clean old welcome messages or not
func GetCleanWelcome(chatID string) int {
	return GetWelcomePrefs(chatID).CleanWelcome
}

// UserClickedButton Mark the user as a human
func UserClickedButton(userID, chatID string) {
	mu := &MutedUser{UserId: userID, ChatId: chatID, ButtonClicked: true}
	SESSION.Save(mu)
}

// HasUserClickedButton Has the user clicked button to unmute themselves
func HasUserClickedButton(userID, chatID string) bool {
	mu := &MutedUser{UserId: userID, ChatId: chatID}
	SESSION.FirstOrInit(mu)
	return mu.ButtonClicked
}

// IsUserHuman Is the user a human
func IsUserHuman(userID, chatID string) bool {
	mu := &MutedUser{UserId: userID, ChatId: chatID}
	return SESSION.First(mu).RowsAffected != 0
}

// SetWelcPref Set whether to welcome or not
func SetWelcPref(chatID string, pref bool) {
	w := &Welcome{ChatId: chatID}
	tx := SESSION.Begin()
	tx.FirstOrCreate(w)
	w.ShouldWelcome = pref
	tx.Save(w)
	tx.Commit()
}

// SetCustomWelcome Set the custom welcome string
func SetCustomWelcome(chatID string, welcome string, buttons []WelcomeButton, welcType int) {
	w := &Welcome{ChatId: chatID}
	if buttons == nil {
		buttons = make([]WelcomeButton, 0)
	}

	tx := SESSION.Begin()
	prevButtons := make([]WelcomeButton, 0)
	tx.Where(&WelcomeButton{ChatId: chatID}).Find(&prevButtons)
	for _, btn := range prevButtons {
		tx.Delete(&btn)
	}

	for _, btn := range buttons {
		tx.Save(&btn)
	}

	tx.FirstOrCreate(w)
	w.CustomWelcome = welcome
	w.WelcomeType = welcType
	tx.Save(w)
	tx.Commit()
}

// GetDelPref Get Whether to delete service messages or not
func GetDelPref(chatID string) bool {
	return GetWelcomePrefs(chatID).DelJoined
}

// SetDelPref Set whether to delete service messages or not
func SetDelPref(chatID string, pref bool) {
	w := &Welcome{ChatId: chatID}
	tx := SESSION.Begin()
	tx.FirstOrCreate(w)
	w.DelJoined = pref
	tx.Save(w)
	tx.Commit()
}

// SetMutePref Set whether to mute users when they join or not
func SetMutePref(chatID string, pref bool) {
	w := &Welcome{ChatId: chatID}
	tx := SESSION.Begin()
	tx.FirstOrCreate(w)
	w.ShouldMute = pref
	tx.Save(w)
	tx.Commit()
}
