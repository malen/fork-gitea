package task

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"code.gitea.io/gitea/modules/context"
	"github.com/xuri/excelize/v2"
)

type MessageJson struct {
	MessageId   string `json:"messageId"`
	MessageText string `json:"messageText"`
}

func GenerateMessageJson(ctx *context.Context) {
	// messageファイル
	// D:/★FM-MAT/trunk/01_受領資料/01_外部設計書/01.全体/12.メッセージ一覧/【暫定版】外部設計_MAT_メッセージ一覧_20220526.xlsx
	const msgExcelPath string = "D:/★FM-MAT/trunk/01_受領資料/01_外部設計書/01.全体/12.メッセージ一覧/【暫定版】外部設計_MAT_メッセージ一覧_20220526.xlsx"

	f, err := excelize.OpenFile(msgExcelPath)
	if err != nil {
		fmt.Printf("open %s, error %v \n", msgExcelPath, err)
	}
	defer func() {
		// Close the spreadsheet.
		if err := f.Close(); err != nil {
			fmt.Printf("close %s, error %v \n", msgExcelPath, err)
		}
	}()

	result := []MessageJson{}
	reg1 := regexp.MustCompile(`MT[0-9]{6}`)
	// 確認メッセージ
	rows1, _ := f.Rows("メッセージ一覧(確認）")
	for rows1.Next() {
		row, _ := rows1.Columns()

		if len(row) >= 5 && reg1.MatchString(row[3]) {
			msg := MessageJson{
				MessageId:   row[3],
				MessageText: row[4],
			}
			result = append(result, msg)
		}
	}

	// 警告メッセージ
	rows2, _ := f.Rows("メッセージ一覧(警告）")
	for rows2.Next() {
		row, _ := rows2.Columns()

		if len(row) >= 5 && reg1.MatchString(row[3]) {
			msg := MessageJson{
				MessageId:   row[3],
				MessageText: row[4],
			}
			result = append(result, msg)
		}
	}

	for _, re := range result {
		fmt.Println(re.MessageId, re.MessageText)
	}

	//array2 := [3]string{"go", "golang", "gopher"}
	// json.MarshalIndent(s, "", "    ")
	b, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Printf("json marshal failed. %v \n", err)
	}

	fmt.Printf("メッセージ件数: %d \n", len(result))
	fmt.Println(string(b))

	ctx.Data["PageIsMiniTask"] = true
	ctx.HTML(http.StatusOK, tplMiniTask)
}
