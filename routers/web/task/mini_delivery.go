package task

import (
	"fmt"
	"net/http"
	"time"
	"strings"

	"code.gitea.io/gitea/modules/base"
	"code.gitea.io/gitea/modules/context"
	"github.com/xuri/excelize/v2"
)

const (
	tplMiniDelivery    base.TplName = "task/delivery"
)

type Task1 struct {
	Value string
	Check bool
}
func Delivery(ctx *context.Context) {

	fmt.Println(ctx.FormString("type"))

	if len(ctx.FormString("type")) == 0 {
		ctx.Data["PageIsMiniTask"] = true
		ctx.HTML(http.StatusOK, tplMiniDelivery)
		return
	}
	// type=1:今週納品予定(先週完了分)
	// type=2:今週完了予定
	// type=3:来週完了予定

	f, err := excelize.OpenFile("D:/★FM-MAT/trunk/00_進捗管理/八丁堀内部管理用スケジュール.xlsx")
    if err != nil {
        fmt.Printf("open %s, error %v", "", err)
        return
    }
    defer func() {
        // Close the spreadsheet.
        if err := f.Close(); err != nil {
            fmt.Printf("close %s, error %v", "", err)
        }
    }()

	rows, err := f.Rows("八丁堀内部スケジュール")
	if err != nil {
		return
	}

	gyomuInfos, _ := ScanFD()

	var taskInfos []TaskInfo
	var kinoId string
	var isBreak bool
	for rows.Next() {
		col, _ := rows.Columns()

		fmt.Printf("xxxxxxxxxxxxxxx%v \n", col)
		// データ有無をチェックする
		if len(col) <= 2 || !isNum(col[1]) {
			continue
		}

		isBreak = false
		for _, gyomuInfo := range gyomuInfos {
			for _, kino := range gyomuInfo.KinoData {
				kinoId = ""
				if kino.Name == convert2(col, 3, false) {
					kinoId = kino.ID
					isBreak = true
					break
				}
			}

			if (isBreak) {
				break
			}
		}

		tskInfo := TaskInfo{
			Kbn: convert2(col, 2, false),
			GyomuKbn: convert2(col, 0, false),
			KinoID: kinoId,
			KinoName: convert2(col, 3, false),
			Phase: convert2(col, 4, false),
			PlanStart: convert2(col, 5, true),
			PlanEnd: convert2(col, 6, true),
			RealStart: convert2(col, 7, true),
			RealEnd: convert2(col, 8, true),
			ReViewPlanStart: convert2(col, 13, true),
			ReViewPlanEnd: convert2(col, 14, true),
			ReViewRealStart: convert2(col, 15, true),
			ReViewRealEnd: convert2(col, 16, true),
			PIC: convert2(col, 9, false),
		}

		taskInfos = append(taskInfos, tskInfo)
	}

	// xiangxi, _ := getAllFiles("D:/★FM-MAT/trunk/02_成果物/01_詳細設計")
	// for _, taskInfo := range taskInfos {

	// 	if len(taskInfo.RealEnd) > 0 && taskInfo.RealEnd != "1899/12/30" {
	// 		if taskInfo.Phase == "内部設計" || taskInfo.Phase == "詳細設計" {
	// 			if taskInfo.Kbn == "API" {
	// 				checkAPIID(xiangxi, taskInfo)
	// 			} else {
	// 				checkPRID(xiangxi, taskInfo)
	// 				checkAPID(xiangxi, taskInfo)
	// 			}
	// 		}
	// 	}
	// }

	// 今週作成予定もの
	var weekTask1 []Task1  // 内部設計
	var weekTask2 []string  // 製造
	var weekTask3 []string  // 単体設計
	var weekTask4 []string  // 単体実施
	// -1：前週、0：当週、1：来週
	w := 0
	if ctx.FormString("type") == "1" {
		w = -1
		ctx.Data["BtnType"] = "1"
	} else if ctx.FormString("type") == "2" {
		w = 0
		ctx.Data["BtnType"] = "2"
	} else {
		w = 1
		ctx.Data["BtnType"] = "3"
	}
	start, end := WeekIntervalTime(w)
	fmt.Printf("---------------------------------------------------------\n")
	fmt.Printf("※報告期間：%s～%s\n", start.Format("2006/01/02"), end.Format("2006/01/02"))
	for _, taskInfo := range taskInfos {									
		if len(taskInfo.PlanEnd) > 0 {
			t, _ := time.Parse("2006/01/02", taskInfo.PlanEnd)

			if t.Before(start) || t.After(end) {
				continue
			}

			//fmt.Printf("taskInfo.RealEnd %s  Phase %s \n", taskInfo.RealEnd, taskInfo.Phase)
			if taskInfo.Phase == "内部設計" || taskInfo.Phase == "詳細設計" {
				if taskInfo.Kbn == "API" {
					expectFileName := fmt.Sprintf("\t詳細設計_MAT_API_SQL仕様((%s)_(%s)).xlsx", taskInfo.KinoID, taskInfo.KinoName)
					task1 := Task1{
						Value: expectFileName,
						Check: false,
					}
					weekTask1 = append(weekTask1, task1)
				} else {
					expectFileName := fmt.Sprintf("\t詳細設計_MAT_PR層ビジネスロジック仕様(%s)_%s.xlsx", taskInfo.KinoID, taskInfo.KinoName)
					task1 := Task1{
						Value: expectFileName,
						Check: false,
					}
					weekTask1 = append(weekTask1, task1)
					expectFileName  = fmt.Sprintf("\t詳細設計_MAT_AP層ビジネスロジック仕様(%s)_%s.xlsx", taskInfo.KinoID, taskInfo.KinoName)
					task1 = Task1{
						Value: expectFileName,
						Check: false,
					}
					weekTask1 = append(weekTask1, task1)
				}
			} else if taskInfo.Phase == "製造" {
				weekTask2 = append(weekTask2, fmt.Sprintf("\t%sのソース一式", taskInfo.KinoName))
			} else if taskInfo.Phase == "単体設計" {
				expectFileName := fmt.Sprintf("\t単体テスト仕様書兼報告書_(%s)_%s.xlsx", taskInfo.KinoID, taskInfo.KinoName)
				weekTask3 = append(weekTask3, expectFileName)
			} else if taskInfo.Phase == "単体テスト" {
				expectFileName := fmt.Sprintf("\t単体テスト仕様書兼報告書_(%s)_%s_テスト結果.xlsx", taskInfo.KinoID, taskInfo.KinoName)
				weekTask4 = append(weekTask4, expectFileName)
			}
		}
		//fmt.Printf("%s,%s,%s,%s,%s,%s,%s\n", taskInfo.Kbn, taskInfo.KinoName, taskInfo.Phase, taskInfo.PlanStart, taskInfo.PlanEnd, taskInfo.ReViewPlanStart, taskInfo.ReViewPlanEnd)
	}
	var build strings.Builder
	build.WriteString("\n内部設計\n")
	//build.WriteString(strings.Join(weekTask1.value, "\n"))
	build.WriteString("\n製造\n")
	build.WriteString(strings.Join(weekTask2, "\n"))
	build.WriteString("\n単体設計\n")
	build.WriteString(strings.Join(weekTask3, "\n"))
	build.WriteString("\n単体実施\n")
	build.WriteString(strings.Join(weekTask4, "\n"))
	fmt.Println(build.String())


	// var weekTask1 []string  // 内部設計
	// var weekTask2 []string  // 製造
	// var weekTask3 []string  // 単体設計
	// var weekTask4 []string  // 単体実施
	ctx.Data["WeekTask1"] = weekTask1
	ctx.Data["WeekTask2"] = weekTask2
	ctx.Data["WeekTask3"] = weekTask3
	ctx.Data["WeekTask4"] = weekTask4
	ctx.Data["PageIsMiniTask"] = true
	ctx.HTML(http.StatusOK, tplMiniDelivery)
}

func getVersionList(filePath string) ([]string) {
	f, err := excelize.OpenFile(filePath)
    if err != nil {
        fmt.Printf("open %s, error %v", "", err)
        return nil
    }
    defer func() {
        if err := f.Close(); err != nil {
            fmt.Printf("close %s, error %v", "", err)
        }
    }()

	rows, err := f.Rows("改版履歴")
	if err != nil {
		return nil
	}

	result := []string{}
	for rows.Next() {
		col, _ := rows.Columns()
		if len(col) <= 2 || !isNum(col[1]) {
			continue
		}

		result = append(result, col[0])
	}

	return result
}