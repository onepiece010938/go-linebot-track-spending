package router

import (
	"fmt"
	"net/http"
)

func DownloadFile(w http.ResponseWriter, r *http.Request) {

	r.ParseForm() //解析url傳遞的參數，對於POST則解析響應包的主體（request body）
	//注意:如果沒有調用ParseForm方法，下面無法獲取表單的數據
	uid := r.Form["uid"]
	fmt.Println(uid)
	file := "./history/" + uid[0] + "/history.csv"
	// file := "./history/U8a75228c06f/history.csv"

	// 設定此 Header 告訴瀏覽器下載檔案。 如果沒設定則會在新的 tab 開啟檔案。
	w.Header().Set("Content-Disposition", "attachment; filename="+file)

	http.ServeFile(w, r, file)
}
