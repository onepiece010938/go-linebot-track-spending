package pkg

import (
	"context"
	"fmt"
	"log"

	"github.com/line/line-bot-sdk-go/linebot"
)

func GuideMessage(event *linebot.Event, client *linebot.Client, ctx context.Context, message *linebot.TextMessage) {
	/* 引導訊息 */
	//要求 新增column時的 引導訊息
	if message.Text == "[request_add_new_column]" {
		fmt.Println(" ")
		// reply message
		if _, err := client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("輸入 新增-項目名稱 來增加你要的項目\n EX:新增-伙食費")).Do(); err != nil {
			log.Println(err.Error())
		}
	}
	//要求 更新 本月餘額上限時的 引導訊息
	if message.Text == "[request_modify_balance_upperbound]" {
		fmt.Println(" ")
		// reply message
		if _, err := client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("輸入 修改餘額上限-$金額 來設定你本月餘額的上限\n EX:修改餘額上限-$45000")).Do(); err != nil {
			log.Println(err.Error())
		}
	}
	//要求 更新 項目餘額上限時的 引導訊息
	if message.Text == "[request_modify_item_upperbound]" {
		fmt.Println(" ")
		// reply message
		if _, err := client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("輸入 修改上限-項目名稱-$金額上限 來設定你項目的餘額上限\n EX:修改上限-伙食費-$6000")).Do(); err != nil {
			log.Println(err.Error())
		}
	}
	// 要求 增加 記帳紀錄時的 引導訊息
	if message.Text == "[request_add_item_detail]" {
		fmt.Println(" ")
		// reply message
		if _, err := client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("輸入 寶-項目欄位-名稱[空格]$金額 來記帳\n EX:寶-伙食費-便當 $100\n")).Do(); err != nil {
			log.Println(err.Error())
		}
	}
	// 要求 刪除 記帳紀錄時的 引導訊息
	if message.Text == "[request_delete_item_detail]" {
		fmt.Println(" ")
		// reply message
		if _, err := client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("輸入 寶-刪除-項目欄位-名稱 來刪除該筆記帳紀錄\n EX:寶-刪除-伙食費-便當\n")).Do(); err != nil {
			log.Println(err.Error())
		}
	}

	// help訊息
	if message.Text == "[help]" {
		fmt.Println(" ")
		ranc := "1.輸入 新增-項目名稱 來增加你要的項目\n EX:新增-伙食費\n"
		rmbu := "2.輸入 修改餘額上限-金額 來設定你本月餘額的上限\n EX:修改餘額上限-$45000\n"
		rmiu := "3.輸入 修改上限-項目名稱-金額上限 來設定你項目的餘額上限\n EX:修改上限-伙食費-$6000\n"
		add := "4.輸入 寶-項目欄位-名稱[空格]$金額 來記帳\n EX:寶-伙食費-便當 $100\n"
		delete := "5.輸入 寶-刪除-項目欄位-名稱 來刪除該筆記帳紀錄\n EX:寶-刪除-伙食費-便當\n"
		new_month := "6.輸入 寶-進入下個月 來將這個月的紀錄存入歷史紀錄,並初始化所有欄位"
		total := ranc + rmbu + rmiu + add + delete + new_month

		// reply message
		if _, err := client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(total)).Do(); err != nil {
			log.Println(err.Error())
		}
	}
}
