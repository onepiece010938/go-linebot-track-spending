package bot

import (
	"go-line-bot/models"
	"go-line-bot/router"
	"log"
	"net/http"

	_ "github.com/joho/godotenv/autoload"
	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/spf13/viper"

	// "io"
	"context"

	"github.com/go-redis/redis/v8"
)

var (
	err error
)

func ServerStart() {
	// writeCsvFile()

	// 建立客戶端
	router.LClient, err = linebot.New(viper.GetString("CHANNEL_SECRET"), viper.GetString("CHANNEL_ACCESS_TOKEN"))

	if err != nil {
		log.Println(err.Error())
	}
	//connect redis
	models.Rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	_, err1 := models.Rdb.Ping(context.Background()).Result()
	if err != nil {
		log.Printf("redis connect get failed.%v", err1)
		return
	}
	log.Printf("redis connect success")

	//初始化 資料庫
	//定義 允許名單資料表
	models.Rdb.HSet(router.Ctx, "allow_user", "line_login_user_id", "ok")

	// bloomFilter(ctx, rdb)

	http.HandleFunc("/download/", router.DownloadFile)

	http.HandleFunc("/callback", router.CallbackHandler)
	log.Printf("start serve on port 5055")
	log.Fatal(http.ListenAndServe(":5055", nil))

}
