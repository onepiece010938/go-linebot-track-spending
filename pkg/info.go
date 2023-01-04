package pkg

import (
	"fmt"

	"github.com/line/line-bot-sdk-go/linebot"
)

func GetUserInfo(event *linebot.Event, client *linebot.Client) (userID string, user_name string, user_column_set string) {
	//抓 userid
	userID = event.Source.UserID
	fmt.Println(userID)
	//抓使用者名稱
	profile, _ := client.GetProfile(userID).Do()
	user_name = profile.DisplayName
	fmt.Println(user_name)
	user_column_set = userID + "user_column_set"

	return userID, user_name, user_column_set
}
