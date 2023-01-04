package router

import (
	"bytes"
	"fmt"
	"go-line-bot/models"
	"go-line-bot/pkg"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/line/line-bot-sdk-go/linebot"
)

func createKeyValuePairs(m map[string]string) string {
	b := new(bytes.Buffer)
	for key, value := range m {
		fmt.Fprintf(b, "%s=\"$%s\"\n", key, value)
	}
	return b.String()
}

func CallbackHandler(w http.ResponseWriter, r *http.Request) {
	client := LClient
	rdb := models.Rdb
	ctx := Ctx
	//清空所有數據庫
	// rdb.FlushAll(ctx).Result()

	// get reauest
	events, err := client.ParseRequest(r)

	if err != nil {
		if err == linebot.ErrInvalidSignature {
			w.WriteHeader(400)
		} else {
			w.WriteHeader(500)
		}

		return
	}

	// writeCsvFile()
	// records := readCsvFile("data.csv")
	// fmt.Println(records)

	/*
		使用者登入 檢查
	*/
	for _, event := range events {

		UserID, user_name, user_column_set := pkg.GetUserInfo(event, client)
		//抓日期
		timeStr := time.Now().Format("2006-01-02 15:04:05")
		dateStr := timeStr[:10] //2022-01-01
		yesterday_date := time.Now().Add(-24 * time.Hour).Format("2006-01-02 15:04:05")[:10]

		fmt.Println(timeStr[8:10])
		fmt.Println(timeStr[:10])

		pkg.Authorization(event, client, rdb, ctx, UserID, user_column_set, dateStr, yesterday_date)

		//判斷是否在名單資料庫內 (要放這是因為放前面就沒機會給使用者輸入是誰了)
		_, err := rdb.HGet(ctx, "allow_user", UserID).Result()
		if err != nil {
			_, err := client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("嗨 "+user_name+" ～，我是機器人-大寶，可能因為第一次使用或系統重啟我現在不認識你,請先告訴我你是誰, 我在決定要不要為你服務 (請回傳 :我是XX)")).Do()
			log.Println(err.Error())
		}

		//從db撈month_check
		month_check, err_mon := rdb.HGet(ctx, UserID, dateStr).Result()
		fmt.Println(month_check, err_mon)
		if err_mon != nil {
			month_check = "no"
		}
		//如果是 當月 1號 要提醒 換新的記帳 不給使用
		if timeStr[8:10] == "01" && month_check != "ok" {
			client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("新的一個月到摟～ 確認好這個月的明細 然後輸入:\n 寶-進入下個月 \n來繼續進行記帳")).Do()
		}

		if event.Type == linebot.EventTypeMessage {
			// 抓使用者的hash table 查餘額
			total_balance, err := rdb.HGet(ctx, UserID, "本月餘額").Result()
			//第一次登入 初始化值
			if err != nil {
				rdb.HSet(ctx, UserID, "本月餘額上限", 30000)
				rdb.HSet(ctx, UserID, "本月餘額", 30000)

				rdb.SAdd(ctx, user_column_set, "伙食費")
				rdb.HSet(ctx, UserID, "伙食費"+"上限", 5000)
				rdb.HSet(ctx, UserID, "伙食費"+"餘額", 5000)
				//創建 紀錄伙食費的 hash table
				rdb.HSet(ctx, UserID+"伙食費", "大寶的愛", "無價")

				rdb.HSet(ctx, UserID, "已分配的餘額", 5000)

				fmt.Println("第一次登入 初始化餘額為30000 本月餘額上限30000")
			}
			fmt.Println("本月餘額: ", total_balance)

			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				// 引導訊息
				pkg.GuideMessage(event, client, ctx, message)

				if message.Text == "[history]" {
					//https://b2b3-125-227-131-83.ngrok.io/download/?uid=1000
					//uid 要等於userid
					url := "https://b2b3-125-227-131-83.ngrok.io/download"
					uid := UserID
					url = url + "/?uid=" + uid
					// reply message
					if _, err = client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(url)).Do(); err != nil {
						log.Println(err.Error())
					}
				}

				//新增column 時的動作
				if strings.Contains(message.Text, "新增-") {
					pkg.AddColumn(event, client, rdb, ctx, message, UserID, user_column_set)

				}

				//修改 本月餘額上限 時的動作
				if strings.Contains(message.Text, "修改餘額上限-") {
					//抓出數字
					new_upper_bound := strings.Split(message.Text, "$")[1]
					//轉int
					new_int_upperbound, _ := strconv.Atoi(new_upper_bound)
					// 檢核判斷
					//抓已分配的 餘額
					allocation_upperbound, _ := rdb.HGet(ctx, UserID, "已分配的餘額").Result()
					int_allocation_upperbound, _ := strconv.Atoi(allocation_upperbound)

					old_upperbound, _ := rdb.HGet(ctx, UserID, "本月餘額上限").Result()
					int_old_upperbound, _ := strconv.Atoi(old_upperbound)

					// 設定過低 (本月餘額上限 不可低於 已分配的餘額)
					if new_int_upperbound < int_allocation_upperbound {
						client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("臭寶 本月餘額上限 不可低於 已分配的餘額, 請嘗試移除其他項目 或是 降低其他項目的餘額上限 ")).Do()
					}
					// 設定低於原本餘額上限 本月餘額要下調
					if (new_int_upperbound < int_old_upperbound) && (new_int_upperbound > int_allocation_upperbound) {
						total_balance, _ := rdb.HGet(ctx, UserID, "本月餘額").Result()
						//轉int計算
						int_total_balance, _ := strconv.Atoi(total_balance)
						new_total_balance := int_total_balance - (int_old_upperbound - new_int_upperbound)
						//重設 DB 本月餘額
						rdb.HSet(ctx, UserID, "本月餘額", new_total_balance)
						//重設 DB 本月餘額上限
						rdb.HSet(ctx, UserID, "本月餘額上限", new_int_upperbound)
						//轉回string
						string_new_upperbound := strconv.Itoa(new_int_upperbound)
						string_new_total_balance := strconv.Itoa(new_total_balance)
						client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("修改 本月餘額上限"+string_new_upperbound+"成功\n同步降低 本月餘額->"+string_new_total_balance)).Do()
					}

					//設定高於原本餘額上限 本月餘額要增加
					if new_int_upperbound > int_old_upperbound {
						total_balance, _ := rdb.HGet(ctx, UserID, "本月餘額").Result()
						int_total_balance, _ := strconv.Atoi(total_balance)
						new_total_balance := int_total_balance + (new_int_upperbound - int_old_upperbound)
						//重設 DB 本月餘額
						rdb.HSet(ctx, UserID, "本月餘額", new_total_balance)
						//重設 DB 本月餘額上限
						rdb.HSet(ctx, UserID, "本月餘額上限", new_int_upperbound)
						//轉回string
						string_new_upperbound := strconv.Itoa(new_int_upperbound)
						string_new_total_balance := strconv.Itoa(new_total_balance)
						client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("修改 本月餘額上限"+string_new_upperbound+"成功\n同步增加 本月餘額->"+string_new_total_balance)).Do()
					}

				}

				//修改 項目 餘額上限 時的動作  修改上限-伙食費-$6000
				if strings.Contains(message.Text, "修改上限-") {
					//抓出項目名稱與金額上限
					tmp := strings.Split(message.Text, "-")
					item_name := tmp[1]
					new_upper_bound := tmp[2]
					new_upper_bound = strings.Split(new_upper_bound, "$")[1]
					// fmt.Println(item_name,new_upper_bound)
					// DB撈舊資料
					old_upper_bound, _ := rdb.HGet(ctx, UserID, item_name+"上限").Result()
					item_balance, _ := rdb.HGet(ctx, UserID, item_name+"餘額").Result()
					//rdb.HSet(ctx, UserID+"伙食費","大寶的愛","無價")
					//轉int
					int_new_upperbound, _ := strconv.Atoi(new_upper_bound)
					int_old_upper_bound, _ := strconv.Atoi(old_upper_bound)
					int_item_balance, _ := strconv.Atoi(item_balance)
					// 檢核判斷
					if int_new_upperbound == 0 {
						client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("指令打錯了 笨寶")).Do()
					}
					if int_new_upperbound == int_old_upper_bound {
						client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("設的值跟原本一樣 87寶")).Do()
					}
					// 使用者調低 項目餘額上限的情況
					if int_new_upperbound < int_old_upper_bound {
						// 目前已被記帳 使用掉多少餘額
						item_balance_used := int_old_upper_bound - int_item_balance
						// 設定過低 (項目餘額上限 不可低於 目前項目已使用的餘額)
						if int_new_upperbound < item_balance_used {
							client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("臭寶 泥這個項目記帳記了太多,用掉太多餘額了 不能設太低")).Do()
						} else {
							// 設定沒有過低  但低於原本項目餘額上限 項目餘額要下調
							//重設 DB item 餘額
							new_int_item_balance := int_item_balance - (int_old_upper_bound - int_new_upperbound)
							rdb.HSet(ctx, UserID, item_name+"餘額", new_int_item_balance)
							//重設 DB item 餘額上限
							rdb.HSet(ctx, UserID, item_name+"上限", int_new_upperbound)
							//轉回string
							string_new_upperbound := strconv.Itoa(int_new_upperbound)
							string_new_total_balance := strconv.Itoa(new_int_item_balance)
							// 同時也要讓 本月的[已分配的餘額]下調
							old_allocation_upperbound, _ := rdb.HGet(ctx, UserID, "已分配的餘額").Result()
							int_old_allocation_upperbound, _ := strconv.Atoi(old_allocation_upperbound)
							new_allocation_upperbound := int_old_allocation_upperbound - (int_old_upper_bound - int_new_upperbound)
							rdb.HSet(ctx, UserID, "已分配的餘額", new_allocation_upperbound) //釋放 餘額上限給 已分配的餘額

							client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("修改 "+item_name+" 餘額上限"+string_new_upperbound+"成功\n同步降低"+item_name+"餘額->"+string_new_total_balance)).Do()
						}
					}
					//使用者調高 項目餘額上限的情況
					if int_new_upperbound > int_old_upper_bound {
						//要首先考量 算本月可分配的餘額 看還剩多少
						allocation_upperbound, _ := rdb.HGet(ctx, UserID, "已分配的餘額").Result()
						int_allocation_upperbound, _ := strconv.Atoi(allocation_upperbound)

						total_balance_upperbound, _ := rdb.HGet(ctx, UserID, "本月餘額上限").Result()
						int_total_balance_upperbound, _ := strconv.Atoi(total_balance_upperbound)
						// 本月還可分配的餘額
						upperbound_for_distribute := int_total_balance_upperbound - int_allocation_upperbound
						upper_bound_diff := int_new_upperbound - int_old_upper_bound

						// 新餘額上限 增加的量 不可超過 還可分配的餘額
						if upper_bound_diff > upperbound_for_distribute {
							string_upperbound_for_distribute := strconv.Itoa(upperbound_for_distribute)
							client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("泥設定的上限超過本月餘額可分配的量, 可分配的餘額剩下:"+string_upperbound_for_distribute)).Do()
						} else {
							// 沒超過本月可分配餘額的話 本月餘額要增加
							new_int_item_balance := upper_bound_diff + int_item_balance
							//重設 DB item 餘額
							rdb.HSet(ctx, UserID, item_name+"餘額", new_int_item_balance)
							//重設 DB item 餘額上限
							rdb.HSet(ctx, UserID, item_name+"上限", int_new_upperbound)
							//轉回string
							string_new_upperbound := strconv.Itoa(int_new_upperbound)
							string_new_total_balance := strconv.Itoa(new_int_item_balance)
							// 同時也要讓 本月的[已分配的餘額]上升
							old_allocation_upperbound, _ := rdb.HGet(ctx, UserID, "已分配的餘額").Result()
							int_old_allocation_upperbound, _ := strconv.Atoi(old_allocation_upperbound)
							new_allocation_upperbound := int_old_allocation_upperbound + (int_new_upperbound - int_old_upper_bound)
							rdb.HSet(ctx, UserID, "已分配的餘額", new_allocation_upperbound) //多佔用  已分配的餘額
							client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("修改 "+item_name+" 餘額上限"+string_new_upperbound+"成功\n同步增加"+item_name+"餘額->"+string_new_total_balance)).Do()
							fmt.Println(upperbound_for_distribute)
						}
					}
				}

				// 新增記帳時的動作 寶-伙食費-便當 $100          要檢核不是做刪除 不然刪除指令也會跑到新增 因為都有 寶-
				if strings.Contains(message.Text, "寶-") && !strings.Contains(message.Text, "寶-刪除-") {
					//抓出項目名稱
					column_name := strings.Split(message.Text, "-")[1]
					//抓出 紀錄名稱 跟金額
					tmp := strings.Split(message.Text, "-")[2]
					add_item_name := strings.Split(tmp, "$")[0]
					//去除左右空白
					add_item_name = strings.TrimSpace(add_item_name)
					add_item_value := strings.Split(tmp, "$")[1]
					// 檢核 項目名稱是否存在
					flag := 0
					all_columns_name, _ := rdb.SMembers(ctx, user_column_set).Result()
					for _, column_name_in_set := range all_columns_name {
						if column_name == column_name_in_set {
							flag = 1
							fmt.Println("in")
						}
					}

					if flag == 0 {
						fmt.Println("項目名稱不存在")
						// reply message
						if _, err = client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("失敗!!!"+column_name+" 不存在 (請確認項目名稱是否正確)")).Do(); err != nil {
							log.Println(err.Error())
						}
					}
					//檢核item_name是否重複 重複要幫加編號
					res, _ := rdb.HExists(ctx, UserID+column_name, add_item_name).Result()
					i := 2
					for { //無限迴圈
						// 舊的key存在了
						if res {
							//後面加編號
							str_i := strconv.Itoa(i)
							if i != 2 {
								add_item_name = strings.Split(add_item_name, "-")[0]
							}
							add_item_name = add_item_name + "-" + str_i
							i++
						}
						//再檢查一次
						res, _ := rdb.HExists(ctx, UserID+column_name, add_item_name).Result()
						//為false 脫離迴圈
						if !res {
							break
						}
					}

					//檢核 可用餘額夠不夠
					// 抓項目可用餘額
					item_balance, _ := rdb.HGet(ctx, UserID, column_name+"餘額").Result()
					// 要計算的 轉int
					int_add_item_value, _ := strconv.Atoi(add_item_value)
					int_item_balance, _ := strconv.Atoi(item_balance)
					if int_add_item_value > int_item_balance {
						client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(column_name+" 餘額不足了!!  "+add_item_value+"超過"+item_balance)).Do()
					} else {
						//記帳紀錄 寫入UserID+column_name 的 hash table
						rdb.HSet(ctx, UserID+column_name, add_item_name, add_item_value)
						// 項目的可用餘額要減這筆
						new_item_balance := int_item_balance - int_add_item_value
						rdb.HSet(ctx, UserID, column_name+"餘額", new_item_balance)
						// 本月總餘額也要減這筆
						total_balance, _ := rdb.HGet(ctx, UserID, "本月餘額").Result()
						int_total_balance, _ := strconv.Atoi(total_balance)
						new_total_balance := int_total_balance - int_add_item_value
						rdb.HSet(ctx, UserID, "本月餘額", new_total_balance)

						// reply message
						if _, err = client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("**"+column_name+"新增 "+add_item_name+" $"+add_item_value+" 成功**")).Do(); err != nil {
							log.Println(err.Error())
						}
						fmt.Println("**" + column_name + "新增 " + add_item_name + "  " + add_item_value + " 成功**")

					}
				}

				// 刪除輸入錯誤的 記帳內容 寶-刪除-伙食費-便當
				if strings.Contains(message.Text, "寶-刪除-") {
					tmp := strings.Split(message.Text, "-")
					//抓出項目名稱
					column_name := tmp[2]
					//抓出 紀錄名稱
					item_name := tmp[3]
					//要處理遇到便當-2的狀況 或是便當-x-x
					if len(tmp) > 4 {
						tmp2 := strings.Split(message.Text, "寶-刪除-"+column_name+"-")
						// fmt.Println(tmp2)
						item_name = tmp2[1]
						// fmt.Println(item_name)
					}
					//去除左右空白
					item_name = strings.TrimSpace(item_name)
					fmt.Println(item_name)

					//要把可用餘額補回去
					//抓 項目餘額
					column_val, _ := rdb.HGet(ctx, UserID, column_name+"餘額").Result()
					//抓 紀錄金額
					item_val, _ := rdb.HGet(ctx, UserID+column_name, item_name).Result()
					// 要計算的 轉int
					int_column_val, _ := strconv.Atoi(column_val)
					int_item_val, _ := strconv.Atoi(item_val)
					recover_column_val := int_column_val + int_item_val
					// 寫入db
					rdb.HSet(ctx, UserID, column_name+"餘額", recover_column_val)
					// 刪除紀錄
					val, _ := rdb.HDel(ctx, UserID+column_name, item_name).Result()
					// fmt.Println(val, err)
					if val == 0 {
						client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("!!"+column_name+"刪除 "+item_name+" 失敗!!\n 請確認 有沒有打錯")).Do()
					}
					client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("**"+column_name+"刪除 "+item_name+" 該筆資料 成功**")).Do()
				}

				if message.Text == "[查詢表單]" {
					//計算總column數
					total_column, err := rdb.SCard(ctx, user_column_set).Result()
					if err != nil {
						panic(err)
					}
					fmt.Printf("總共有%v個欄位(default 1個本月餘額)", total_column)
					// 用來存所有bubble的slice 拿來回傳Carousel
					collect_template := []*linebot.BubbleContainer{}

					//取出所有元素
					all_columns_name, err := rdb.SMembers(ctx, user_column_set).Result()
					if err != nil {
						panic(err)
					}
					fmt.Println(all_columns_name[0])
					fmt.Printf("Datatype of all_columns_name : %T\n", all_columns_name)

					for i := 0; i < int(total_column); i++ {
						print("")
						i_column := all_columns_name[i]
						if i_column == "本月餘額" {
							total_balance, err := rdb.HGet(ctx, UserID, "本月餘額").Result()
							if err != nil {
								fmt.Println(err)
								panic(err)
							}
							balance_upper_bound, err := rdb.HGet(ctx, UserID, "本月餘額上限").Result()
							if err != nil {
								fmt.Println(err)
								panic(err)
							}
							allocation_upperbound, _ := rdb.HGet(ctx, UserID, "已分配的餘額").Result()
							//創見餘額的template
							template := &linebot.BubbleContainer{
								Type: linebot.FlexContainerTypeBubble,
								Body: &linebot.BoxComponent{
									Type:   linebot.FlexComponentTypeBox,
									Layout: linebot.FlexBoxLayoutTypeVertical,

									Contents: []linebot.FlexComponent{
										&linebot.TextComponent{
											Type:  linebot.FlexComponentTypeText,
											Text:  "本月餘額上限: " + balance_upper_bound,
											Style: "italic",
										},
										&linebot.TextComponent{
											Type:  linebot.FlexComponentTypeText,
											Text:  "已分配餘額:" + allocation_upperbound,
											Style: "italic",
										},
										&linebot.TextComponent{
											Type:   linebot.FlexComponentTypeText,
											Text:   "\n本月餘額: " + total_balance + "\n",
											Color:  "#77A88D",
											Weight: "bold",
											Size:   "xl",
										},
										&linebot.ButtonComponent{
											/*
												return &PostbackAction{
													Label:       label,
													Data:        data,
													Text:        text,
													DisplayText: displayText,
												}*/
											Type:   "button",
											Style:  "primary",
											Color:  "#B4D3AA",
											Margin: "xxl",
											Height: "sm",
											// Action:linebot.NewURIAction("Go to line.me", "https://line.me"),
											// Action:linebot.NewPostbackAction("Say hello1", "hello 1", "", "傳我出去"),
											// Action:linebot.NewPostbackAction("我是按鈕名稱", "我不知道這啥", "傳我出去", ""),
											Action: linebot.NewMessageAction("修改餘額上限", "[request_modify_balance_upperbound]"),
											// AdjustMode: "shrink-to-fit",
										},
										&linebot.ButtonComponent{

											Type:   "button",
											Style:  "primary",
											Color:  "#B4D3AA",
											Margin: "xxl",
											Height: "sm",
											Action: linebot.NewMessageAction("重新 查詢表單", "[查詢表單]"),
										},
									},
								},
							}
							//把本月餘額的flex放進去  插到最前面
							collect_template = append([]*linebot.BubbleContainer{template}, collect_template...)
						} else {
							//撈該項目全部的記帳紀錄
							detail, err := rdb.HGetAll(ctx, UserID+i_column).Result()
							detail_string := ""
							if err == nil {
								detail_string = createKeyValuePairs(detail)
							} else {
								fmt.Println(err)
								detail_string = "無紀錄"
							}
							item_upperbound, _ := rdb.HGet(ctx, UserID, i_column+"上限").Result()
							item_balance, _ := rdb.HGet(ctx, UserID, i_column+"餘額").Result()
							//創建 bubble的template
							template := &linebot.BubbleContainer{
								Type: linebot.FlexContainerTypeBubble,
								Body: &linebot.BoxComponent{
									Type:   linebot.FlexComponentTypeBox,
									Layout: linebot.FlexBoxLayoutTypeVertical,
									Contents: []linebot.FlexComponent{
										&linebot.TextComponent{
											Type:  linebot.FlexComponentTypeText,
											Text:  i_column + " 餘額上限:" + item_upperbound,
											Style: "italic",
											Wrap:  true,
										},
										&linebot.TextComponent{
											Type:   linebot.FlexComponentTypeText,
											Text:   i_column + " 可用餘額:" + item_balance,
											Color:  "#77A88D",
											Weight: "bold",
											Size:   "xl",
											Wrap:   true,
										},
										&linebot.TextComponent{
											Type:  linebot.FlexComponentTypeText,
											Text:  detail_string,
											Color: "#000080",
											Size:  "sm",
											Wrap:  true,
										},
										&linebot.ButtonComponent{
											Type:   "button",
											Style:  "primary",
											Color:  "#B4D3AA",
											Height: "sm",
											Margin: "md", // none, xs, sm, md, lg, xl, or xxl
											// Action:linebot.NewURIAction("Go to line.me", "https://line.me"),
											// Action:linebot.NewPostbackAction("Say hello1", "hello 1", "", "傳我出去"),
											// Action:linebot.NewPostbackAction("我是按鈕名稱", "我不知道這啥", "傳我出去", ""),
											Action: linebot.NewMessageAction("修改["+i_column+"]上限", "[request_modify_item_upperbound]"),
											// AdjustMode: "shrink-to-fit",
										},
										&linebot.ButtonComponent{
											Type:   "button",
											Style:  "primary",
											Color:  "#B4D3AA",
											Height: "sm",
											Margin: "md",
											// Action:linebot.NewURIAction("Go to line.me", "https://line.me"),
											// Action:linebot.NewPostbackAction("Say hello1", "hello 1", "", "傳我出去"),
											// Action:linebot.NewPostbackAction("我是按鈕名稱", "我不知道這啥", "傳我出去", ""),
											Action: linebot.NewMessageAction("新增["+i_column+"] 記帳紀錄", "[request_add_item_detail]"),
											// AdjustMode: "shrink-to-fit",
										},
										&linebot.ButtonComponent{
											Type:   "button",
											Style:  "primary",
											Color:  "#B4D3AA",
											Height: "sm",
											Margin: "md",
											// Action:linebot.NewURIAction("Go to line.me", "https://line.me"),
											// Action:linebot.NewPostbackAction("Say hello1", "hello 1", "", "傳我出去"),
											// Action:linebot.NewPostbackAction("我是按鈕名稱", "我不知道這啥", "傳我出去", ""),
											Action: linebot.NewMessageAction("刪除["+i_column+"]的某則記帳紀錄", "[request_delete_item_detail]"),
											// AdjustMode: "shrink-to-fit",
										},
									},
								},
							}

							collect_template = append(collect_template, template)

						}

					}
					// 出for後 collect_template 已經做好了 把他弄成carousel候傳出去
					contents := &linebot.CarouselContainer{
						Type:     linebot.FlexContainerTypeCarousel,
						Contents: collect_template,
					}
					if _, err := client.ReplyMessage(
						event.ReplyToken,
						linebot.NewFlexMessage("Flex message alt text", contents),
					).Do(); err != nil {
						return
					}
					print(collect_template)
					fmt.Println("給出全部的bubble 要從DB確認現在有多少column然後看要作幾個bubble")
				}

				if message.Text == "[check_balance]" {

					// jsonString:=get_json()
					// contents, err := linebot.UnmarshalFlexMessageJSON([]byte(jsonString))
					// if err != nil {
					// 	return
					// }
					// if _, err := client.ReplyMessage(
					// 	event.ReplyToken,
					// 	linebot.NewFlexMessage("Flex message alt text", contents),
					// ).Do(); err != nil {
					// 	return
					// }
					// {
					//   "type": "carousel",
					//   "contents": [
					//     {
					//       "type": "bubble",
					//       "body": {
					//         "type": "box",
					//         "layout": "vertical",
					//         "contents": [
					//           {
					//             "type": "text",
					//             "text": "First bubble"
					//           }
					//         ]
					//       }
					//     },
					//     {
					//       "type": "bubble",
					//       "body": {
					//         "type": "box",
					//         "layout": "vertical",
					//         "contents": [
					//           {
					//             "type": "text",
					//             "text": "Second bubble"
					//           }
					//         ]
					//       }
					//     }
					//   ]
					// }

					contents := &linebot.CarouselContainer{
						Type: linebot.FlexContainerTypeCarousel,
						Contents: []*linebot.BubbleContainer{
							{
								Type: linebot.FlexContainerTypeBubble,
								Body: &linebot.BoxComponent{
									Type:   linebot.FlexComponentTypeBox,
									Layout: linebot.FlexBoxLayoutTypeVertical,
									Contents: []linebot.FlexComponent{
										&linebot.TextComponent{
											Type: linebot.FlexComponentTypeText,
											Text: "First bubble",
										},
									},
								},
							},
							{
								Type: linebot.FlexContainerTypeBubble,
								Body: &linebot.BoxComponent{
									Type:   linebot.FlexComponentTypeBox,
									Layout: linebot.FlexBoxLayoutTypeVertical,

									Contents: []linebot.FlexComponent{
										&linebot.TextComponent{
											Type: linebot.FlexComponentTypeText,
											Text: "AAAAA",
										},
										&linebot.ButtonComponent{
											/*
												return &PostbackAction{
													Label:       label,
													Data:        data,
													Text:        text,
													DisplayText: displayText,
												}*/
											Type: "button",
											// Action:linebot.NewURIAction("Go to line.me", "https://line.me"),
											// Action:linebot.NewPostbackAction("Say hello1", "hello 1", "", "傳我出去"),
											Action: linebot.NewPostbackAction("我是按鈕名稱", "我不知道這啥", "傳我出去", ""),
											// Action:linebot.NewMessageAction("Say message", "Rice=米"),
											// AdjustMode: "shrink-to-fit",
										},
									},
								},
							},
						},
					}
					if _, err := client.ReplyMessage(
						event.ReplyToken,
						linebot.NewFlexMessage("Flex message alt text", contents),
					).Do(); err != nil {
						return
					}
				}
				// reply message
				if _, err = client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(message.Text+"自動回話")).Do(); err != nil {
					log.Println(err.Error())
				}
			}
		}

	}

}

/*
func get_json()string{
	jsonString := `{
		"type": "bubble",
		"hero": {
		  "type": "image",
		  "url": "https://scdn.line-apps.com/n/channel_devcenter/img/fx/01_1_cafe.png",
		  "size": "full",
		  "aspectRatio": "20:13",
		  "aspectMode": "cover",
		  "action": {
			"type": "uri",
			"uri": "http://linecorp.com/"
		  }
		},
		"body": {
		  "type": "box",
		  "layout": "vertical",
		  "contents": [
			{
			  "type": "text",
			  "text": "Brown Cafe",
			  "weight": "bold",
			  "size": "xl"
			},
			{
			  "type": "box",
			  "layout": "baseline",
			  "margin": "md",
			  "contents": [
				{
				  "type": "icon",
				  "size": "sm",
				  "url": "https://scdn.line-apps.com/n/channel_devcenter/img/fx/review_gold_star_28.png"
				},
				{
				  "type": "icon",
				  "size": "sm",
				  "url": "https://scdn.line-apps.com/n/channel_devcenter/img/fx/review_gold_star_28.png"
				},
				{
				  "type": "icon",
				  "size": "sm",
				  "url": "https://scdn.line-apps.com/n/channel_devcenter/img/fx/review_gold_star_28.png"
				},
				{
				  "type": "icon",
				  "size": "sm",
				  "url": "https://scdn.line-apps.com/n/channel_devcenter/img/fx/review_gold_star_28.png"
				},
				{
				  "type": "icon",
				  "size": "sm",
				  "url": "https://scdn.line-apps.com/n/channel_devcenter/img/fx/review_gray_star_28.png"
				},
				{
				  "type": "text",
				  "text": "4.0",
				  "size": "sm",
				  "color": "#999999",
				  "margin": "md",
				  "flex": 0
				}
			  ]
			},
			{
			  "type": "box",
			  "layout": "vertical",
			  "margin": "lg",
			  "spacing": "sm",
			  "contents": [
				{
				  "type": "box",
				  "layout": "baseline",
				  "spacing": "sm",
				  "contents": [
					{
					  "type": "text",
					  "text": "Place",
					  "color": "#aaaaaa",
					  "size": "sm",
					  "flex": 1
					},
					{
					  "type": "text",
					  "text": "Miraina Tower, 4-1-6 Shinjuku, Tokyo",
					  "wrap": true,
					  "color": "#666666",
					  "size": "sm",
					  "flex": 5
					}
				  ]
				},
				{
				  "type": "box",
				  "layout": "baseline",
				  "spacing": "sm",
				  "contents": [
					{
					  "type": "text",
					  "text": "Time",
					  "color": "#aaaaaa",
					  "size": "sm",
					  "flex": 1
					},
					{
					  "type": "text",
					  "text": "10:00 - 23:00",
					  "wrap": true,
					  "color": "#666666",
					  "size": "sm",
					  "flex": 5
					}
				  ]
				}
			  ]
			}
		  ]
		},
		"footer": {
		  "type": "box",
		  "layout": "vertical",
		  "spacing": "sm",
		  "contents": [
			{
			  "type": "button",
			  "style": "link",
			  "height": "sm",
			  "action": {
				"type": "uri",
				"label": "CALL",
				"uri": "https://linecorp.com"
			  }
			},
			{
			  "type": "button",
			  "style": "link",
			  "height": "sm",
			  "action": {
				"type": "uri",
				"label": "WEBSITE",
				"uri": "https://linecorp.com",
				"altUri": {
				  "desktop": "https://line.me/ja/download"
				}
			  }
			},
			{
			  "type": "spacer",
			  "size": "sm"
			}
		  ],
		  "flex": 0
		}
	  }`
	return  jsonString
}
*/
/*
紅色 = #FF0000 = RGB(255, 0, 0)

綠色 = #008000 = RGB(1, 128, 0)

藍色 = #0000FF = RGB(0, 0, 255)

白色 = #FFFFFF = RGB(255,255,255)

象牙色 = #FFFFF0 = RGB(255, 255, 240)

黑色 = #000000 = RGB(0, 0, 0)

灰色 = #808080 = RGB(128, 128, 128)

銀 = #C0C0C0 = RGB(192, 192, 192)

黃色 = #FFFF00 = RGB(255, 255, 0)

紫色 = #800080 = RGB(128, 0, 128)

橙色 = FFA500 = RGB(255, 165, 0)

栗色 = #800000 = RGB(128, 0, 0)

紫紅色 = #FF00FF = RGB(255, 0, 255)

石灰 = #00FF00 = RGB(0, 255, 0)

水色 = #00FFFF = RGB(0, 255, 255)

青色 = #008080 = RGB(0, 128, 128)

橄欖色 = #808000 = RGB(128, 128, 0)

海軍 = #000080 = RGB(0, 0, 128)
*/
