package main

import (
	_ "github.com/joho/godotenv/autoload"
	"github.com/line/line-bot-sdk-go/linebot"
	"log"
	"net/http"
	"os"
	"fmt"
	"strconv"
	"bytes"
	"time"
	// "io"
	"encoding/csv"
	"context"
    "github.com/go-redis/redis/v8"
	"strings"
)

var (
	client *linebot.Client
	err    error
	ctx = context.Background()
	rdb *redis.Client
)

func main() {
	// writeCsvFile()
	// 建立客戶端
	client, err = linebot.New(os.Getenv("CHANNEL_SECRET"), os.Getenv("CHANNEL_ACCESS_TOKEN"))

	if err != nil {
		log.Println(err.Error())
	}
	//connect redis
	rdb = redis.NewClient(&redis.Options{
        Addr:     "localhost:6666",
        Password: "", // no password set
        DB:       0,  // use default DB
    })

	_,err1 := rdb.Ping(context.Background()).Result()
	if err != nil{
		log.Printf("redis connect get failed.%v",err1)
		return
	}
	log.Printf("redis connect success")

	//初始化 資料庫
	//定義 允許名單資料表
	rdb.HSet(ctx, "allow_user", "line_login_user_id", "ok")


	// bloomFilter(ctx, rdb)


	http.HandleFunc("/download/", downloadFile)

	http.HandleFunc("/callback", callbackHandler)

	log.Fatal(http.ListenAndServe(":5055", nil))


	
}
//判斷字串內是否含有某些字串
// func contains(s []string, str string) bool {
// 	for _, v := range s {
// 		if v == str {
// 			return true
// 		}
// 	}

// 	return false
// }

/*
func readCsvFile(filePath string) [][]string {
    f, err := os.Open(filePath)
    if err != nil {
        log.Fatal("Unable to read input file " + filePath, err)
    }
    defer f.Close()

    csvReader := csv.NewReader(f)
    records, err := csvReader.ReadAll()
    if err != nil {
        log.Fatal("Unable to parse file as CSV for " + filePath, err)
    }

    return records
}*/


/*
func writeCsvFile() {

    // records := [][]string{
	// 	{"2022-01-03"},
    //     {"first_name", "last_name", "occupation"},
    //     {"John", "Doe", "gardener"},
    //     {"Lucy", "Smith", "teacher"},
    //     {"Brian", "Bethamy", "programmer"},
    // }
	records := [][]string{}
	// records[0][0] = "aaa"
	// records[3][3] = "aa2"
	array1 := []string{"sss"}
	tmp :="qqqqq"
	array1 = append(array1,tmp)
	array2 := []string{"ttttt"}
	records = append(records, array1,array2)
	// records= append(records ,"1", "2")

    f, err := os.Create("data.csv")
	if err != nil {
        log.Fatal("Unable to write input file ")
    }
    defer f.Close()

    if err != nil {

        log.Fatalln("failed to open file", err)
    }
	// fmt.Println(records[0][0])
	// fmt.Println(records[3][1])

    w := csv.NewWriter(f)
    err = w.WriteAll(records) // calls Flush internally

    if err != nil {
        log.Fatal(err)
    }
}*/

func createKeyValuePairs(m map[string]string) string {
    b := new(bytes.Buffer)
    for key, value := range m {
        fmt.Fprintf(b, "%s=\"$%s\"\n", key, value)
    }
    return b.String()
}

func downloadFile(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()//解析url傳遞的參數，對於POST則解析響應包的主體（request body）
	//注意:如果沒有調用ParseForm方法，下面無法獲取表單的數據
	uid := r.Form["uid"]
	fmt.Println(uid)
	file := "./history/"+uid[0]+"/history.csv"
	// file := "./history/U8a75228c06f/history.csv"

	// 設定此 Header 告訴瀏覽器下載檔案。 如果沒設定則會在新的 tab 開啟檔案。
	w.Header().Set("Content-Disposition", "attachment; filename="+file)

	http.ServeFile(w, r, file)
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}





func callbackHandler(w http.ResponseWriter, r *http.Request) {
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
		//抓 userid
		UserID := event.Source.UserID
		fmt.Println(UserID)
		//抓使用者名稱
		profile, _ := client.GetProfile(UserID).Do()
		user_name := profile.DisplayName
		fmt.Println(user_name)
		user_column_set := UserID+"user_column_set"
		//抓日期
		timeStr:=time.Now().Format("2006-01-02 15:04:05") 
		dateStr := timeStr[:10]//2022-01-01
		yesterday_date := time.Now().Add(-24*time.Hour).Format("2006-01-02 15:04:05")[:10] 

		fmt.Println(timeStr[8:10]) 
		fmt.Println(timeStr[:10]) 





		
		// 判斷是否在名單內之後，如果回答正確，把使用者id加入允許名單
		if event.Type == linebot.EventTypeMessage {
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				if message.Text =="sudo init"{
					rdb.HSet(ctx, UserID,"本月餘額上限",30000)
					rdb.HSet(ctx, UserID,"本月餘額",30000)
					
					rdb.SAdd(ctx, user_column_set, "伙食費")
					rdb.HSet(ctx, UserID,"伙食費"+"上限",5000)
					rdb.HSet(ctx, UserID,"伙食費"+"餘額",5000)
					//創建 紀錄伙食費的 hash table
					rdb.HSet(ctx, UserID+"伙食費","大寶的愛","無價")

					rdb.HSet(ctx, UserID,"已分配的餘額",5000)
					client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("初始化ok")).Do()
						
				}
				// reply message
				if message.Text =="我是小寶"{
					//加入allow_user的key
					rdb.HSet(ctx, "allow_user", UserID, "ok")
					//第一次登入的初始化設定 創建名稱是userid的hash table
					rdb.HMSet(ctx, UserID, map[string]interface{}{"name":"小寶", "user_id":UserID})
					//創建要存 使用者自定義的欄位名稱的set
					user_column_set := UserID+"user_column_set"
					rdb.SAdd(ctx, user_column_set, "本月餘額")
					client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("原來是我滴愛寶哇!! 泥好我滴愛寶")).Do()
					
				}
				if message.Text =="我是大寶"{
					//加入allow_user的key
					rdb.HSet(ctx, "allow_user", UserID, "ok")
					//第一次登入的初始化設定 創建名稱是userid的hash table
					rdb.HMSet(ctx, UserID, map[string]interface{}{"name":"大寶", "user_id":UserID})
					//創建要存 使用者自定義的欄位名稱的set
					user_column_set := UserID+"user_column_set"
					rdb.SAdd(ctx, user_column_set, "本月餘額")
					client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("歡迎 admin 大寶")).Do()
					
				}
				if message.Text =="寶-進入下個月"{
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
					row1 = append(row1,yesterday_date) 
					total_balance, _ := rdb.HGet(ctx, UserID, "本月餘額").Result()
					row1 = append(row1,"本月總餘額:",total_balance) 
					// 存進records
					records = append(records,row1) 

					//剩下幾個透過迴圈做
					
					row2:= []string{}
					row3:= []string{}
					row4:= []string{}

					for i:=0;i<int(total_column);i++{
						i_column := all_columns_name[i]
						if i_column=="本月餘額"{
							continue
						}
						row2 =  append(row2,i_column) 
						fmt.Println("row2",row2)
						detail,err := rdb.HGetAll(ctx, UserID+i_column).Result()
						detail_string:=""
						fmt.Println(detail_string)
						if err == nil {
							detail_string = createKeyValuePairs(detail)
							fmt.Println(detail_string)
						}else{
							fmt.Println(err)
							detail_string="無紀錄"
						}
						row3 =  append(row3,detail_string)
						fmt.Println("row3",row3)

						item_balance,_ := rdb.HGet(ctx, UserID,i_column+"餘額").Result()
						row4=  append(row4,i_column+"餘額: $"+item_balance+"\n")
						fmt.Println("row4",row4)
					}
					records = append(records,row2,row3,row4) 
					fmt.Println(records)
					// ---開始寫入 csv---

					save_dir :="./history/"+UserID+"/"
					save_file := save_dir +"history.csv"

					// 路徑不存在 建立路徑
					bo_result,_ := PathExists(save_dir)
					if !bo_result{
						err:=os.Mkdir(save_dir, os.ModePerm)
						fmt.Println(err)
					}

					file, er := os.Open(save_file)
					
					// 如果文件不存在，建立文件
					if er != nil && os.IsNotExist(er) {
						
						file,_=os.Create(save_file)
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
					for i:=0;i<int(total_column);i++{
						i_column := all_columns_name[i]
						if i_column=="本月餘額"{
							//要把本月餘額還原成 原本上定的上限
							t_upper,_:=rdb.HGet(ctx, UserID,"本月餘額上限").Result()
							rdb.HSet(ctx, UserID,"本月餘額",t_upper)
							continue
						}
						//刪除記帳資訊
						n, _ := rdb.Unlink(ctx, UserID+i_column).Result()
						//初始化 該項目 記帳資訊
						rdb.HSet(ctx, UserID+i_column,"大寶的愛","無價")
						fmt.Println("Unlink",n)
						//要把 各項目的餘額還原成原本設定的上限
						i_upper,_:=rdb.HGet(ctx, UserID,i_column+"上限").Result()
						rdb.HSet(ctx, UserID,i_column+"餘額",i_upper)

					}
					//全部做完 這個月的flag設為ok
					rdb.HSet(ctx,UserID, dateStr,"ok")
					client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(dateStr+" 收到～愛寶 繼續記帳gogo")).Do()
				}



				}
			}
		

		//判斷是否在名單資料庫內 (要放這是因為放前面就沒機會給使用者輸入是誰了)
		_, err := rdb.HGet(ctx, "allow_user", UserID).Result()
		if err != nil {
			_, err := client.ReplyMessage(event.ReplyToken,linebot.NewTextMessage("嗨 "+user_name+" ～，我是機器人-大寶，可能因為第一次使用或系統重啟我現在不認識你,請先告訴我你是誰, 我在決定要不要為你服務 (請回傳 :我是XX)")).Do()
			log.Println(err.Error())
		}


		//從db撈month_check
		month_check, err_mon := rdb.HGet(ctx,UserID, dateStr).Result()
		fmt.Println(month_check,err_mon)
		if err_mon != nil {
				month_check="no"
			}
		//如果是 當月 1號 要提醒 換新的記帳 不給使用
		if timeStr[8:10]=="01" && month_check!="ok"{
			client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("新的一個月到摟～ 確認好這個月的明細 然後輸入:\n 寶-進入下個月 \n來繼續進行記帳")).Do()
		}

		

		if event.Type == linebot.EventTypeMessage {
			// 抓使用者的hash table 查餘額
			total_balance, err := rdb.HGet(ctx, UserID, "本月餘額").Result()
			//第一次登入 初始化值
			if err != nil {
				rdb.HSet(ctx, UserID,"本月餘額上限",30000)
				rdb.HSet(ctx, UserID,"本月餘額",30000)
				
				rdb.SAdd(ctx, user_column_set, "伙食費")
				rdb.HSet(ctx, UserID,"伙食費"+"上限",5000)
				rdb.HSet(ctx, UserID,"伙食費"+"餘額",5000)
				//創建 紀錄伙食費的 hash table
				rdb.HSet(ctx, UserID+"伙食費","大寶的愛","無價")

				rdb.HSet(ctx, UserID,"已分配的餘額",5000)
				
				fmt.Println("第一次登入 初始化餘額為30000 本月餘額上限30000")
			}
			fmt.Println("本月餘額: ",total_balance)

			
			
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				/* 引導訊息 */
				//要求 新增column時的 引導訊息
				if message.Text =="[request_add_new_column]"{
					fmt.Println(" ")
					// reply message
					if _, err = client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("輸入 新增-項目名稱 來增加你要的項目\n EX:新增-伙食費")).Do(); err != nil {
						log.Println(err.Error())
					}
				}
				//要求 更新 本月餘額上限時的 引導訊息
				if message.Text =="[request_modify_balance_upperbound]"{
					fmt.Println(" ")
					// reply message
					if _, err = client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("輸入 修改餘額上限-$金額 來設定你本月餘額的上限\n EX:修改餘額上限-$45000")).Do(); err != nil {
						log.Println(err.Error())
					}
				}
				//要求 更新 項目餘額上限時的 引導訊息
				if message.Text =="[request_modify_item_upperbound]"{
					fmt.Println(" ")
					// reply message
					if _, err = client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("輸入 修改上限-項目名稱-$金額上限 來設定你項目的餘額上限\n EX:修改上限-伙食費-$6000")).Do(); err != nil {
						log.Println(err.Error())
					}
				}
				// 要求 增加 記帳紀錄時的 引導訊息
				if message.Text =="[request_add_item_detail]"{
					fmt.Println(" ")
					// reply message
					if _, err = client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("輸入 寶-項目欄位-名稱[空格]$金額 來記帳\n EX:寶-伙食費-便當 $100\n")).Do(); err != nil {
						log.Println(err.Error())
					}
				}
				// 要求 刪除 記帳紀錄時的 引導訊息
				if message.Text =="[request_delete_item_detail]"{
					fmt.Println(" ")
					// reply message
					if _, err = client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("輸入 寶-刪除-項目欄位-名稱 來刪除該筆記帳紀錄\n EX:寶-刪除-伙食費-便當\n")).Do(); err != nil {
						log.Println(err.Error())
					}
				}
				
				// help訊息
				if message.Text =="[help]"{
					fmt.Println(" ")
					ranc :="1.輸入 新增-項目名稱 來增加你要的項目\n EX:新增-伙食費\n"
					rmbu := "2.輸入 修改餘額上限-金額 來設定你本月餘額的上限\n EX:修改餘額上限-$45000\n"
					rmiu :="3.輸入 修改上限-項目名稱-金額上限 來設定你項目的餘額上限\n EX:修改上限-伙食費-$6000\n"
					add :="4.輸入 寶-項目欄位-名稱[空格]$金額 來記帳\n EX:寶-伙食費-便當 $100\n"
					delete := "5.輸入 寶-刪除-項目欄位-名稱 來刪除該筆記帳紀錄\n EX:寶-刪除-伙食費-便當\n"
					new_month := "6.輸入 寶-進入下個月 來將這個月的紀錄存入歷史紀錄,並初始化所有欄位"
					total := ranc+rmbu+rmiu+add+delete+new_month


					// reply message
					if _, err = client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(total)).Do(); err != nil {
						log.Println(err.Error())
					}
				}

				if message.Text =="[history]"{
					//https://b2b3-125-227-131-83.ngrok.io/download/?uid=1000
					//uid 要等於userid
					url:="https://b2b3-125-227-131-83.ngrok.io/download"
					uid:=UserID
					url = url+"/?uid="+uid
					// reply message
					if _, err = client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(url)).Do(); err != nil {
						log.Println(err.Error())
					}
				}


				//新增column 時的動作
				if strings.Contains(message.Text, "新增-"){
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
						if _, err = client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("失敗!!!"+column_name+" 已經存在 (請確認項目名稱是否正確)")).Do(); err != nil {
							log.Println(err.Error())
						}
					}
					if flag == 0{
						//寫入DB的 hash table跟set
						rdb.HSet(ctx, UserID,column_name,0)
						rdb.SAdd(ctx, user_column_set, column_name)

						rdb.HSet(ctx, UserID+column_name,"大寶的愛","無價")
						rdb.HSet(ctx, UserID,column_name+"上限",0)
						rdb.HSet(ctx, UserID,column_name+"餘額",0)

						// reply message
						if _, err = client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("**新增 "+column_name+" 成功**")).Do(); err != nil {
							log.Println(err.Error())
						}
						fmt.Println("寫入 ",column_name," 成功")
					}
					
				}

				//修改 本月餘額上限 時的動作
				if strings.Contains(message.Text, "修改餘額上限-"){
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
					if new_int_upperbound < int_allocation_upperbound{
						client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("臭寶 本月餘額上限 不可低於 已分配的餘額, 請嘗試移除其他項目 或是 降低其他項目的餘額上限 ")).Do()
					}
					// 設定低於原本餘額上限 本月餘額要下調
					if (new_int_upperbound < int_old_upperbound) && (new_int_upperbound > int_allocation_upperbound){
						total_balance, _ := rdb.HGet(ctx, UserID, "本月餘額").Result()
						//轉int計算
						int_total_balance,_ := strconv.Atoi(total_balance)
						new_total_balance := int_total_balance - (int_old_upperbound - new_int_upperbound)
						//重設 DB 本月餘額
						rdb.HSet(ctx, UserID,"本月餘額",new_total_balance)
						//重設 DB 本月餘額上限
						rdb.HSet(ctx, UserID,"本月餘額上限",new_int_upperbound)
						//轉回string
						string_new_upperbound:=strconv.Itoa(new_int_upperbound)
						string_new_total_balance:=strconv.Itoa(new_total_balance)
						client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("修改 本月餘額上限"+string_new_upperbound +"成功\n同步降低 本月餘額->"+string_new_total_balance)).Do()
					}

					//設定高於原本餘額上限 本月餘額要增加
					if new_int_upperbound > int_old_upperbound{
						total_balance, _ := rdb.HGet(ctx, UserID, "本月餘額").Result()
						int_total_balance,_ := strconv.Atoi(total_balance)
						new_total_balance := int_total_balance + (new_int_upperbound-int_old_upperbound)
						//重設 DB 本月餘額
						rdb.HSet(ctx, UserID,"本月餘額",new_total_balance)
						//重設 DB 本月餘額上限
						rdb.HSet(ctx, UserID,"本月餘額上限",new_int_upperbound)
						//轉回string
						string_new_upperbound:=strconv.Itoa(new_int_upperbound)
						string_new_total_balance:=strconv.Itoa(new_total_balance)
						client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("修改 本月餘額上限"+string_new_upperbound +"成功\n同步增加 本月餘額->"+string_new_total_balance)).Do()
					}

				}

				//修改 項目 餘額上限 時的動作  修改上限-伙食費-$6000
				if strings.Contains(message.Text, "修改上限-"){
					//抓出項目名稱與金額上限
					tmp := strings.Split(message.Text, "-")
					item_name := tmp[1]
					new_upper_bound := tmp[2]
					new_upper_bound = strings.Split(new_upper_bound, "$")[1]
					// fmt.Println(item_name,new_upper_bound)
					// DB撈舊資料
					old_upper_bound,_ := rdb.HGet(ctx, UserID,item_name+"上限").Result()
					item_balance,_ := rdb.HGet(ctx, UserID,item_name+"餘額").Result()
					//rdb.HSet(ctx, UserID+"伙食費","大寶的愛","無價")
					//轉int
					int_new_upperbound, _ := strconv.Atoi(new_upper_bound)
					int_old_upper_bound, _ := strconv.Atoi(old_upper_bound)
					int_item_balance, _ := strconv.Atoi(item_balance)
					// 檢核判斷
					if int_new_upperbound == 0{
						client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("指令打錯了 笨寶")).Do()
					}
					if int_new_upperbound == int_old_upper_bound{
						client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("設的值跟原本一樣 87寶")).Do()
					}
					// 使用者調低 項目餘額上限的情況
					if int_new_upperbound <int_old_upper_bound{
						// 目前已被記帳 使用掉多少餘額
						item_balance_used:=int_old_upper_bound - int_item_balance
						// 設定過低 (項目餘額上限 不可低於 目前項目已使用的餘額)
						if int_new_upperbound < item_balance_used{
							client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("臭寶 泥這個項目記帳記了太多,用掉太多餘額了 不能設太低")).Do()
						}else{
							// 設定沒有過低  但低於原本項目餘額上限 項目餘額要下調
							//重設 DB item 餘額
							new_int_item_balance:=int_item_balance-(int_old_upper_bound-int_new_upperbound)
							rdb.HSet(ctx, UserID,item_name+"餘額",new_int_item_balance)
							//重設 DB item 餘額上限
							rdb.HSet(ctx, UserID,item_name+"上限",int_new_upperbound)
							//轉回string
							string_new_upperbound:=strconv.Itoa(int_new_upperbound)
							string_new_total_balance:=strconv.Itoa(new_int_item_balance)
							// 同時也要讓 本月的[已分配的餘額]下調
							old_allocation_upperbound, _ := rdb.HGet(ctx, UserID, "已分配的餘額").Result()
							int_old_allocation_upperbound, _ := strconv.Atoi(old_allocation_upperbound)
							new_allocation_upperbound := int_old_allocation_upperbound-(int_old_upper_bound-int_new_upperbound)
							rdb.HSet(ctx, UserID,"已分配的餘額",new_allocation_upperbound) //釋放 餘額上限給 已分配的餘額

							client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("修改 "+item_name+" 餘額上限"+string_new_upperbound +"成功\n同步降低"+item_name+"餘額->"+string_new_total_balance)).Do()
						}
					}
					//使用者調高 項目餘額上限的情況 
					if int_new_upperbound >int_old_upper_bound{
						//要首先考量 算本月可分配的餘額 看還剩多少 
						allocation_upperbound, _ := rdb.HGet(ctx, UserID, "已分配的餘額").Result()
						int_allocation_upperbound, _ := strconv.Atoi(allocation_upperbound)

						total_balance_upperbound, _ := rdb.HGet(ctx, UserID, "本月餘額上限").Result()
						int_total_balance_upperbound, _ := strconv.Atoi(total_balance_upperbound)
						// 本月還可分配的餘額
						upperbound_for_distribute := int_total_balance_upperbound-int_allocation_upperbound
						upper_bound_diff:=int_new_upperbound - int_old_upper_bound
			
						// 新餘額上限 增加的量 不可超過 還可分配的餘額
						if upper_bound_diff > upperbound_for_distribute{
							string_upperbound_for_distribute:=strconv.Itoa(upperbound_for_distribute)
							client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("泥設定的上限超過本月餘額可分配的量, 可分配的餘額剩下:"+string_upperbound_for_distribute)).Do()
						}else{
							// 沒超過本月可分配餘額的話 本月餘額要增加
							new_int_item_balance:=upper_bound_diff+int_item_balance
							//重設 DB item 餘額
							rdb.HSet(ctx, UserID,item_name+"餘額",new_int_item_balance)
							//重設 DB item 餘額上限
							rdb.HSet(ctx, UserID,item_name+"上限",int_new_upperbound)
							//轉回string
							string_new_upperbound:=strconv.Itoa(int_new_upperbound)
							string_new_total_balance:=strconv.Itoa(new_int_item_balance)
							// 同時也要讓 本月的[已分配的餘額]上升
							old_allocation_upperbound, _ := rdb.HGet(ctx, UserID, "已分配的餘額").Result()
							int_old_allocation_upperbound, _ := strconv.Atoi(old_allocation_upperbound)
							new_allocation_upperbound := int_old_allocation_upperbound+(int_new_upperbound-int_old_upper_bound)
							rdb.HSet(ctx, UserID,"已分配的餘額",new_allocation_upperbound) //多佔用  已分配的餘額
							client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("修改 "+item_name+" 餘額上限"+string_new_upperbound +"成功\n同步增加"+item_name+"餘額->"+string_new_total_balance)).Do()
							fmt.Println(upperbound_for_distribute)
						}
					}
				}

				// 新增記帳時的動作 寶-伙食費-便當 $100          要檢核不是做刪除 不然刪除指令也會跑到新增 因為都有 寶-
				if strings.Contains(message.Text, "寶-")&& !strings.Contains(message.Text, "寶-刪除-"){
					//抓出項目名稱
					column_name := strings.Split(message.Text, "-")[1]
					//抓出 紀錄名稱 跟金額
					tmp := strings.Split(message.Text, "-")[2]
					add_item_name :=  strings.Split(tmp, "$")[0]
					//去除左右空白
					add_item_name = strings.TrimSpace(add_item_name)
					add_item_value :=  strings.Split(tmp, "$")[1]
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
					res,_:=rdb.HExists(ctx, UserID+column_name,add_item_name).Result()
					i := 2
					for { //無限迴圈
						// 舊的key存在了
						if res{
							//後面加編號
							str_i := strconv.Itoa(i)
							if i != 2 {
								add_item_name = strings.Split(add_item_name, "-")[0]
							}
							add_item_name = add_item_name +"-"+str_i
							i++
						}
						//再檢查一次
						res,_:=rdb.HExists(ctx, UserID+column_name,add_item_name).Result()
						//為false 脫離迴圈
						if !res{
							break
						}	
					}

					//檢核 可用餘額夠不夠
					// 抓項目可用餘額
					item_balance,_ :=rdb.HGet(ctx, UserID,column_name+"餘額").Result()
					// 要計算的 轉int
					int_add_item_value,_ := strconv.Atoi(add_item_value)
					int_item_balance,_ :=strconv.Atoi(item_balance)
					if int_add_item_value > int_item_balance{
						client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(column_name+" 餘額不足了!!  "+add_item_value+"超過"+item_balance)).Do()
					}else{
						//記帳紀錄 寫入UserID+column_name 的 hash table
						rdb.HSet(ctx, UserID+column_name,add_item_name,add_item_value)
						// 項目的可用餘額要減這筆
						new_item_balance := int_item_balance - int_add_item_value
						rdb.HSet(ctx, UserID,column_name+"餘額",new_item_balance)
						// 本月總餘額也要減這筆
						total_balance, _ := rdb.HGet(ctx, UserID, "本月餘額").Result()
						int_total_balance,_ := strconv.Atoi(total_balance)
						new_total_balance := int_total_balance - int_add_item_value
						rdb.HSet(ctx, UserID, "本月餘額",new_total_balance)

						// reply message
						if _, err = client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("**"+column_name+"新增 "+add_item_name+" $"+add_item_value+" 成功**")).Do(); err != nil {
							log.Println(err.Error())
						}
						fmt.Println("**"+column_name+"新增 "+add_item_name+"  "+add_item_value+" 成功**")

					}
				}
 
				// 刪除輸入錯誤的 記帳內容 寶-刪除-伙食費-便當
				if strings.Contains(message.Text, "寶-刪除-"){
					tmp :=strings.Split(message.Text, "-")
					//抓出項目名稱
					column_name := tmp[2]
					//抓出 紀錄名稱
					item_name := tmp[3]
					//要處理遇到便當-2的狀況 或是便當-x-x
					if len(tmp)>4{
						tmp2:=strings.Split(message.Text, "寶-刪除-"+column_name+"-")
						// fmt.Println(tmp2)
						item_name = tmp2[1]
						// fmt.Println(item_name)
					}
					//去除左右空白
					item_name = strings.TrimSpace(item_name)
					fmt.Println(item_name)


					//要把可用餘額補回去
					//抓 項目餘額
					column_val,_ := rdb.HGet(ctx, UserID,column_name+"餘額").Result()
					//抓 紀錄金額
					item_val,_ := rdb.HGet(ctx, UserID+column_name, item_name).Result()
					// 要計算的 轉int
					int_column_val,_ := strconv.Atoi(column_val)
					int_item_val,_ :=strconv.Atoi(item_val)
					recover_column_val := int_column_val + int_item_val
					// 寫入db
					rdb.HSet(ctx, UserID,column_name+"餘額",recover_column_val)
					// 刪除紀錄
					val, _ := rdb.HDel(ctx, UserID+column_name, item_name).Result()
					// fmt.Println(val, err)
					if val == 0 {
						client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("!!"+column_name+"刪除 "+item_name+" 失敗!!\n 請確認 有沒有打錯")).Do()
					}
					client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("**"+column_name+"刪除 "+item_name+" 該筆資料 成功**")).Do()
				}

				
				
				if message.Text =="[查詢表單]"{
					//計算總column數
					total_column, err := rdb.SCard(ctx, user_column_set).Result()
					if err != nil {
						panic(err)
					}
					fmt.Println("總共有%v個欄位(default 1個本月餘額)",total_column)
					// 用來存所有bubble的slice 拿來回傳Carousel
					collect_template:=[]*linebot.BubbleContainer{}

					//取出所有元素
					all_columns_name, err := rdb.SMembers(ctx, user_column_set).Result()
					if err != nil {
						panic(err)
					}
					fmt.Println(all_columns_name[0])
					fmt.Printf("Datatype of all_columns_name : %T\n", all_columns_name)

					for i:=0;i<int(total_column);i++{
						print("")
						i_column := all_columns_name[i]
						if i_column=="本月餘額"{
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
											Type: linebot.FlexComponentTypeText,
											Text: "本月餘額上限: "+ balance_upper_bound,
											Style:"italic",
										},
										&linebot.TextComponent{
											Type: linebot.FlexComponentTypeText,
											Text: "已分配餘額:"+ allocation_upperbound,
											Style:"italic",
										},
										&linebot.TextComponent{
											Type: linebot.FlexComponentTypeText,
											Text: "\n本月餘額: "+total_balance+"\n",
											Color:"#77A88D",
											Weight: "bold",
											Size: "xl",
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
											Style: "primary",
											Color:"#B4D3AA",
											Margin:"xxl",
											Height:"sm",
											// Action:linebot.NewURIAction("Go to line.me", "https://line.me"),
											// Action:linebot.NewPostbackAction("Say hello1", "hello 1", "", "傳我出去"),
											// Action:linebot.NewPostbackAction("我是按鈕名稱", "我不知道這啥", "傳我出去", ""),
											Action:linebot.NewMessageAction("修改餘額上限", "[request_modify_balance_upperbound]"),
											// AdjustMode: "shrink-to-fit",
										},
										&linebot.ButtonComponent{

											Type: "button",
											Style: "primary",
											Color:"#B4D3AA",
											Margin:"xxl",
											Height:"sm",
											Action:linebot.NewMessageAction("重新 查詢表單", "[查詢表單]"),
										},
									},
								},
							}
							//把本月餘額的flex放進去  插到最前面
							collect_template=append([]*linebot.BubbleContainer{template},collect_template...)
						}else{
							//撈該項目全部的記帳紀錄
							detail,err := rdb.HGetAll(ctx, UserID+i_column).Result()
							detail_string:=""
							if err == nil {
								detail_string = createKeyValuePairs(detail)
							}else{
								fmt.Println(err)
								detail_string="無紀錄"
							}
							item_upperbound, _ := rdb.HGet(ctx, UserID,i_column+"上限").Result()
							item_balance, _ := rdb.HGet(ctx, UserID,i_column+"餘額").Result()
							//創建 bubble的template
							template := &linebot.BubbleContainer{
								Type: linebot.FlexContainerTypeBubble,
								Body: &linebot.BoxComponent{
									Type:   linebot.FlexComponentTypeBox,
									Layout: linebot.FlexBoxLayoutTypeVertical,
									Contents: []linebot.FlexComponent{
										&linebot.TextComponent{
											Type: linebot.FlexComponentTypeText,
											Text: i_column+" 餘額上限:"+item_upperbound,
											Style:"italic",
											Wrap: true,
										},
										&linebot.TextComponent{
											Type: linebot.FlexComponentTypeText,
											Text: i_column+ " 可用餘額:"+item_balance,
											Color:"#77A88D",
											Weight: "bold",
											Size: "xl",
											Wrap: true,
										},
										&linebot.TextComponent{
											Type: linebot.FlexComponentTypeText,
											Text: detail_string,
											Color:"#000080",
											Size: "sm",
											Wrap: true,
										},
										&linebot.ButtonComponent{
											Type: "button",
											Style: "primary",
											Color:"#B4D3AA",
											Height:"sm",
											Margin:"md", // none, xs, sm, md, lg, xl, or xxl
											// Action:linebot.NewURIAction("Go to line.me", "https://line.me"),
											// Action:linebot.NewPostbackAction("Say hello1", "hello 1", "", "傳我出去"),
											// Action:linebot.NewPostbackAction("我是按鈕名稱", "我不知道這啥", "傳我出去", ""),
											Action:linebot.NewMessageAction("修改["+i_column+"]上限", "[request_modify_item_upperbound]"),
											// AdjustMode: "shrink-to-fit",
										},
										&linebot.ButtonComponent{
											Type: "button",
											Style: "primary",
											Color:"#B4D3AA",
											Height:"sm",
											Margin:"md",
											// Action:linebot.NewURIAction("Go to line.me", "https://line.me"),
											// Action:linebot.NewPostbackAction("Say hello1", "hello 1", "", "傳我出去"),
											// Action:linebot.NewPostbackAction("我是按鈕名稱", "我不知道這啥", "傳我出去", ""),
											Action:linebot.NewMessageAction("新增["+i_column+"] 記帳紀錄", "[request_add_item_detail]"),
											// AdjustMode: "shrink-to-fit",
										},
										&linebot.ButtonComponent{
											Type: "button",
											Style: "primary",
											Color:"#B4D3AA",
											Height:"sm",
											Margin:"md",
											// Action:linebot.NewURIAction("Go to line.me", "https://line.me"),
											// Action:linebot.NewPostbackAction("Say hello1", "hello 1", "", "傳我出去"),
											// Action:linebot.NewPostbackAction("我是按鈕名稱", "我不知道這啥", "傳我出去", ""),
											Action:linebot.NewMessageAction("刪除["+i_column+"]的某則記帳紀錄", "[request_delete_item_detail]"),
											// AdjustMode: "shrink-to-fit",
										},
									},
								},
							}

							collect_template=append(collect_template,template)

						}


						
					}
					// 出for後 collect_template 已經做好了 把他弄成carousel候傳出去
					contents := &linebot.CarouselContainer{
						Type: linebot.FlexContainerTypeCarousel,
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
				
				if message.Text =="[check_balance]"{

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
											Action:linebot.NewPostbackAction("我是按鈕名稱", "我不知道這啥", "傳我出去", ""),
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

/*	redis指令

	// err := rdb.Set(ctx, "key", "666666", 0).Err()
    // if err != nil {
    //     panic(err)
    // }

    // val, err := rdb.Get(ctx, "key").Result()
    // if err != nil {
    //     panic(err)
    // }
    // fmt.Println("key", val)

    // val2, err := rdb.Get(ctx, "key2").Result()
    // if err == redis.Nil {
    //     fmt.Println("key2 does not exist")
    // } else if err != nil {
    //     panic(err)
    // } else {
    //     fmt.Println("key2", val2)
    // }
    // Output: key value
    // key2 does not exist

	//set hash 適合儲存結構  
	// rdb.HSet(ctx, "user", "key1", "value1", "key2", "value2")
	// rdb.Del(ctx, "user")
	// rdb.HMSet(ctx, "user", map[string]interface{}{"name":"kevin", "age": 27, "address":"北京"})

	// //HGet():获取某个元素
	// address, err := rdb.HGet(ctx, "user", "address").Result()
	// if err != nil {
	// 	fmt.Println("addreersss")
	// 	fmt.Println(err)
	// 	panic(err)
	// }
	// fmt.Println(address)

	// //HGetAll():获取全部元素
	// user, err := rdb.HGetAll(ctx, "user").Result()
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(user)

	// //HDel():删除某个元素
	// res, err := rdb.HDel(ctx, "user", "name", "age").Result()
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(res)

*/