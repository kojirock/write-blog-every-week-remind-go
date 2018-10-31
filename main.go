package main

import (
	"context"
	"os"

	"./config"
	"./database"
	"./date"
	"./message"
	"./rss"
	"./slack"
	"github.com/aws/aws-lambda-go/lambda"
)

/**
 * Main.
 */
func main() {
	executeFunction := os.Getenv("GOLANG_EXECUTE_FUNCTION")
	if executeFunction == "reminder" {
		lambda.Start(blogReminder)
		// blogReminder()
	} else if executeFunction == "register" {
		lambda.Start(blogRegister)
	} else if executeFunction == "result" {
		lambda.Start(blogResult)
		// blogResult()
	} else {
		panic("環境変数 GOLANG_EXECUTE_FUNCTION が取得出来ないか、期待した値ではありません。")
	}
}

/**
 * ブログのリマインダーロジックを実行
 */
func blogReminder() {
	const targetHour int = 15
	thisMonday := date.GetThisMonday(targetHour)
	configData := config.GetConfigData()
	allMemberData := database.FindByMemberData(configData)
	targetUserList := rss.FindTargetUserList(allMemberData, thisMonday)
	sendText := message.MakeReminderSendText(targetUserList)
	database.UpdateRequireCount(configData, allMemberData, targetUserList)
	slack.SendMessage(configData, sendText)
	// fmt.Println(sendText)
}

/**
 * ブログの登録ロジックを実行
 */
func blogRegister(_ context.Context, rawParams interface{}) (interface{}, error) {
	configData := config.GetConfigData()
	envToken := os.Getenv("SLACK_TOKEN")
	params, err := slack.ParseSlackParams(rawParams)
	if err != nil {
		return "スラックのパラメータが取得できませんでした。 error: " + err.Error(), nil
	}
	if envToken != params.Token {
		return "トークンの不一致", nil
	}
	userData := database.FindByPK(configData, params.UserID)
	if userData.UserID != "" {
		return "あなたのブログはすでに登録済みです feedURL: " + userData.FeedURL, nil
	}
	database.CreateUser(configData, params)

	return "ブログを登録しました。これからは妥協は許しませんよ。", nil
}

/**
 * ブログ書けたかどうか通知のロジックを実行
 */
func blogResult() {
	const targetHour int = 0
	lastWeekMonday := date.GetLastWeekMonday(targetHour)
	configData := config.GetConfigData()
	allMemberData := database.FindByMemberData(configData)
	targetUserList := rss.FindTargetUserList(allMemberData, lastWeekMonday)
	targetUserList = database.ResetRequireCount(configData, allMemberData, targetUserList)
	sendText := message.MakeResultSendText(targetUserList)
	slack.SendMessage(configData, sendText)
	// fmt.Println(sendText)
}