package pkg

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/line/line-bot-sdk-go/linebot"
)

func Authorization(event *linebot.Event, client *linebot.Client, rdb *redis.Client, ctx context.Context, UserID string, user_column_set string, dateStr string, yesterday_date string) {
	if event.Type == linebot.EventTypeMessage {
		switch message := event.Message.(type) {
		case *linebot.TextMessage:
			if message.Text == "sudo init" {
				rdb.HSet(ctx, UserID, "本月餘額上限", 30000)
				rdb.HSet(ctx, UserID, "本月餘額", 30000)

				rdb.SAdd(ctx, user_column_set, "伙食費")
				rdb.HSet(ctx, UserID, "伙食費"+"上限", 5000)
				rdb.HSet(ctx, UserID, "伙食費"+"餘額", 5000)
				//創建 紀錄伙食費的 hash table
				rdb.HSet(ctx, UserID+"伙食費", "大寶的愛", "無價")

				rdb.HSet(ctx, UserID, "已分配的餘額", 5000)
				client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("初始化ok")).Do()

			}
			// reply message
			if message.Text == "我是小寶" {
				//加入allow_user的key
				rdb.HSet(ctx, "allow_user", UserID, "ok")
				//第一次登入的初始化設定 創建名稱是userid的hash table
				rdb.HMSet(ctx, UserID, map[string]interface{}{"name": "小寶", "user_id": UserID})
				//創建要存 使用者自定義的欄位名稱的set
				user_column_set := UserID + "user_column_set"
				rdb.SAdd(ctx, user_column_set, "本月餘額")
				client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("原來是我滴愛寶哇!! 泥好我滴愛寶")).Do()

			}
			if message.Text == "我是大寶" {
				//加入allow_user的key
				rdb.HSet(ctx, "allow_user", UserID, "ok")
				//第一次登入的初始化設定 創建名稱是userid的hash table
				rdb.HMSet(ctx, UserID, map[string]interface{}{"name": "大寶", "user_id": UserID})
				//創建要存 使用者自定義的欄位名稱的set
				user_column_set := UserID + "user_column_set"
				rdb.SAdd(ctx, user_column_set, "本月餘額")
				client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("歡迎 admin 大寶")).Do()

			}
			if message.Text == "寶-進入下個月" {
				//初始化下個月的資料 同時寫進csv

				//計算總column數
				total_column, _ := rdb.SCard(ctx, user_column_set).Result()
				//取出所有元素
				all_columns_name, _ := rdb.SMembers(ctx, user_column_set).Result()
				//創建二維陣列 來存進csv
				records := [][]string{}
				//存成
				// 2021-12-31 ,"本月餘額",本月餘額
				// "項目1" , "項目2","項目3"
				//  xx:$10 , yy:$10, zz:$10
				// "餘額:"＄xx,"餘額:"＄xx,"餘額:"＄xx

				//先做好第一個row
				row1 := []string{}
				row1 = append(row1, yesterday_date)
				total_balance, _ := rdb.HGet(ctx, UserID, "本月餘額").Result()
				row1 = append(row1, "本月總餘額:", total_balance)
				// 存進records
				records = append(records, row1)

				//剩下幾個透過迴圈做

				row2 := []string{}
				row3 := []string{}
				row4 := []string{}

				for i := 0; i < int(total_column); i++ {
					i_column := all_columns_name[i]
					if i_column == "本月餘額" {
						continue
					}
					row2 = append(row2, i_column)
					fmt.Println("row2", row2)
					detail, err := rdb.HGetAll(ctx, UserID+i_column).Result()
					detail_string := ""
					fmt.Println(detail_string)
					if err == nil {
						detail_string = createKeyValuePairs(detail)
						fmt.Println(detail_string)
					} else {
						fmt.Println(err)
						detail_string = "無紀錄"
					}
					row3 = append(row3, detail_string)
					fmt.Println("row3", row3)

					item_balance, _ := rdb.HGet(ctx, UserID, i_column+"餘額").Result()
					row4 = append(row4, i_column+"餘額: $"+item_balance+"\n")
					fmt.Println("row4", row4)
				}
				records = append(records, row2, row3, row4)
				fmt.Println(records)
				// ---開始寫入 csv---

				save_dir := "./history/" + UserID + "/"
				save_file := save_dir + "history.csv"

				// 路徑不存在 建立路徑
				bo_result, _ := pathExists(save_dir)
				if !bo_result {
					err := os.Mkdir(save_dir, os.ModePerm)
					fmt.Println(err)
				}

				file, er := os.Open(save_file)

				// 如果文件不存在，建立文件
				if er != nil && os.IsNotExist(er) {

					file, _ = os.Create(save_file)
				}
				file.Close()
				// 這樣開 每次都會清空文件 (只有第一次開能用)
				// f, err := os.Create(save_file)
				// 這樣開 就可以在後面追加(linux用法)(要先有檔案可以開)
				f, err := os.OpenFile(save_file, os.O_APPEND|os.O_RDWR, 0666)
				if err != nil {
					log.Fatal("Unable to write input file ")
				}
				defer f.Close()

				if err != nil {

					log.Fatalln("failed to open file", err)
				}
				// fmt.Println(records[0][0])
				// fmt.Println(records[3][1])
				f.WriteString("\xEF\xBB\xBF") // 寫入UTF-8 BOM，防止中文亂碼
				w := csv.NewWriter(f)
				err = w.WriteAll(records) // calls Flush internally

				if err != nil {
					log.Fatal(err)
				}
				// 寫完準備初始化
				for i := 0; i < int(total_column); i++ {
					i_column := all_columns_name[i]
					if i_column == "本月餘額" {
						//要把本月餘額還原成 原本上定的上限
						t_upper, _ := rdb.HGet(ctx, UserID, "本月餘額上限").Result()
						rdb.HSet(ctx, UserID, "本月餘額", t_upper)
						continue
					}
					//刪除記帳資訊
					n, _ := rdb.Unlink(ctx, UserID+i_column).Result()
					//初始化 該項目 記帳資訊
					rdb.HSet(ctx, UserID+i_column, "大寶的愛", "無價")
					fmt.Println("Unlink", n)
					//要把 各項目的餘額還原成原本設定的上限
					i_upper, _ := rdb.HGet(ctx, UserID, i_column+"上限").Result()
					rdb.HSet(ctx, UserID, i_column+"餘額", i_upper)

				}
				//全部做完 這個月的flag設為ok
				rdb.HSet(ctx, UserID, dateStr, "ok")
				client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(dateStr+" 收到～愛寶 繼續記帳gogo")).Do()
			}

		}
	}
}

func createKeyValuePairs(m map[string]string) string {
	b := new(bytes.Buffer)
	for key, value := range m {
		fmt.Fprintf(b, "%s=\"$%s\"\n", key, value)
	}
	return b.String()
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
