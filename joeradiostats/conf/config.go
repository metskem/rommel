package conf

import (
	"log"
	"os"
)

const DatabaseURL = "file:joestats.db"
const CreateTablesFile = "resources/sql/create-tables.sql"

var (
	// Variables to identify the build
	CommitHash string
	VersionTag string
	BuildTime  string

	BotToken         = os.Getenv("BOT_TOKEN")
	DebugStr         = os.Getenv("DEBUG")
	ChromeDriverPath = os.Getenv("CHROME_DRIVER_PATH")
	Debug            bool
)

func EnvironmentComplete() {
	envComplete := true

	if len(BotToken) == 0 {
		log.Print("missing envvar \"BOT_TOKEN\"")
		envComplete = false
	}

	Debug = false
	if DebugStr == "true" {
		Debug = true
	}

	if ChromeDriverPath == "" {
		ChromeDriverPath = "/opt/homebrew/bin/chromedriver"
	}

	if !envComplete {
		log.Fatal("one or more envvars missing, aborting...")
	}
}
