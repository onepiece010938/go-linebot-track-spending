package pkg

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/line/line-bot-sdk-go/linebot"
)

func AddColumn(event *linebot.Event, client *linebot.Client, rdb *redis.Client, ctx context.Context, message *linebot.TextMessage, UserID string, user_column_set string) {
	//抓出欄位名稱
	column_name := strings.Split(message.Text, "-")[1]
	//去除左右空白
	column_name = strings.TrimSpace(column_name)
	// 檢核 項目名稱是否存在
	flag := 0
	all_columns_name, _ := rdb.SMembers(ctx, user_column_set).Result()
	for _, column_name_in_set := range all_columns_name {
		if column_name == column_name_in_set {
			flag = 1
			fmt.Println("in")
		}
	}

	if flag == 1 {
		fmt.Println("項目名稱已存在")
		// reply message
		if _, err := client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("失敗!!!"+column_name+" 已經存在 (請確認項目名稱是否正確)")).Do(); err != nil {
			log.Println(err.Error())
		}
	}
	if flag == 0 {
		//寫入DB的 hash table跟set
		rdb.HSet(ctx, UserID, column_name, 0)
		rdb.SAdd(ctx, user_column_set, column_name)

		rdb.HSet(ctx, UserID+column_name, "大寶的愛", "無價")
		rdb.HSet(ctx, UserID, column_name+"上限", 0)
		rdb.HSet(ctx, UserID, column_name+"餘額", 0)

		// reply message
		if _, err := client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("**新增 "+column_name+" 成功**")).Do(); err != nil {
			log.Println(err.Error())
		}
		fmt.Println("寫入 ", column_name, " 成功")
	}
}
