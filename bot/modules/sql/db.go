package sql

import (
	"log"

	"github.com/anandpskerala/Martha/bot"
	"github.com/anandpskerala/Martha/bot/modules/utils/error_handling"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/lib/pq"
)

var SESSION *gorm.DB

func init() {
	conn, err := pq.ParseURL(bot.BotConfig.SqlUri)
	error_handling.FatalError(err)

	db, err := gorm.Open("postgres", conn)
	if err != nil {
            panic("failed to connect database")
        }
        //defer db.Close()

	if bot.BotConfig.DebugMode == "True" {
		SESSION = db.Debug()
		log.Println("Using database in debug mode.")
	} else {
		SESSION = db
	}

        db.DB().SetMaxIdleConns(10)

      

	log.Println("[INFO][Database] Database connected")

	// Create tables if they don't exist
        SESSION.AutoMigrate(&User{}, &Chat{}, &Rules{}, &Welcome{}, &WelcomeButton{}, &MutedUser{}, &FedChat{}, &Federation{}, &FedAdmin{}, &FedBan{}, &Note{}, &Button{}, &Warns{}, &WarnFilters{}, &WarnSettings{}, &BlackListFilters{}, &Connect{})

	log.Println("Auto-migrated database schema")

}
