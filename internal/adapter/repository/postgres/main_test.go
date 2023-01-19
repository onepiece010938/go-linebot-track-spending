package postgres

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/spf13/viper"
)

// create testStore for all postgres test
var testStore Store

func TestMain(m *testing.M) {
	viper.SetConfigFile("config/go-line-bot.yaml")
	postgres_test_dns := viper.GetString("postgres.test_dsn")
	var err error
	testDB, err := sql.Open("postgres", postgres_test_dns)
	if err != nil {
		log.Fatal("postger connect fail!:", err)
	}
	// For tx
	testStore = NewStore(testDB)
	os.Exit(m.Run())
}
