package task

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"
	"os"
	"encoding/json"

	"code.gitea.io/gitea/modules/base"
	"code.gitea.io/gitea/modules/context"
	//"code.gitea.io/gitea/modules/timeutil"
	task_model "code.gitea.io/gitea/models/task"

	"github.com/xuri/excelize/v2"

	// "xorm.io/xorm"
	// "xorm.io/xorm/schemas"

	_ "github.com/mattn/go-oci8"
	"xorm.io/xorm"
)

const (
	tplMiniTask    base.TplName = "task/minitask"
)

// MiniTask is the minitask page
func MiniTask(ctx *context.Context) {
	var tasks []*task_model.MiniTask

	var err error
	if tasks, err = task_model.GetAllTasks(); err != nil {
		ctx.ServerError("xxxxxxxxxxxxxxxxx: ", err)
		return
	}

	for _, task := range tasks {
		fmt.Printf("xxxxxxxxxxxxxxxxxxxxx%s", task.ID)
	}


	ctx.Data["PageIsMiniTask"] = true
	ctx.HTML(http.StatusOK, tplMiniTask)
}

type TaskInfo struct {
    Kbn		  string `json:"kbn"`        // 大区分 1：画面 2：API
	GyomuKbn  string `json:"gyomu_kbn"`        // 業務名
	KinoID    string `json:"kino_id"`        // 機能ID
	KinoName  string `json:"kino_name"`        // 機能名
	Phase     string `json:"phase"`        // 内部設計、製造、単体設計、単体実施
	PlanStart string `json:"plan_start"`        // 予定開始日
	PlanEnd   string `json:"plan_end"`        // 予定終了日
	RealStart string `json:"real_start"`        // 実績開始日
	RealEnd   string `json:"real_end"`        // 実績終了日
	ReViewPlanStart string `json:"review_plan_start"`  // レ予定開始日
	ReViewPlanEnd   string `json:"review_plan_end"`  // レ予定終了日
	ReViewRealStart string `json:"review_real_start"`  // レ実績開始日
	ReViewRealEnd   string `json:"review_real_end"`  // レ実績終了日
	PIC       string `json:"pic"`        // 担当者
}

// スケジュールからタスク情報を読み込む
func ReadTask(ctx *context.Context) {
    f, err := excelize.OpenFile("D:/★FM-MAT/trunk/00_進捗管理/bak/MAT_オフショア要員_スケジュール.xlsx")
    if err != nil {
        fmt.Printf("open %s, error %v", "", err)
        return
    }
    defer func() {
        // Close the spreadsheet.
        if err := f.Close(); err != nil {
            fmt.Println(err)
        }
    }()

	rows, err := f.Rows("スケジュール_最新")
	if err != nil {
		fmt.Printf("スケジュール_最新 が見つかりません。")
		return
	}


	var taskInfos []TaskInfo
	for rows.Next() {
		row, _ := rows.Columns()

		reg1 := regexp.MustCompile(`.*八丁堀.*`)
		if reg1 == nil {
			continue
		} else {

			if  reg1.MatchString(row[3]) {
				tskInfo := TaskInfo{
					Kbn: row[1],
					GyomuKbn: row[0],
					KinoName: row[2],
					Phase: "内部設計",
					PlanStart: convert(row, 4),
					PlanEnd: convert(row, 5),
					ReViewPlanStart: convert(row, 6),
					ReViewPlanEnd: convert(row, 7),
				}

				taskInfos = append(taskInfos, tskInfo)
				//fmt.Printf("内部設計:%s,%s,%s,%s,%s,%s,%s,%s\n", row[0], row[1], row[2], row[3], row[4], row[5], row[6], row[7])
			} 
			
			if reg1.MatchString(row[13]) {
				tskInfo1 := TaskInfo{
					Kbn: row[1],
					GyomuKbn: row[0],
					KinoName: row[2],
					Phase: "製造",
					PlanStart: convert(row, 14),
					PlanEnd: convert(row, 15),
					ReViewPlanStart: convert(row, 16),
					ReViewPlanEnd: convert(row, 17),
				}

				taskInfos = append(taskInfos, tskInfo1)
				//fmt.Printf("製造:%s,%s,%s,%s,%s,%s,%s,%s\n", row[0], row[1], row[2], row[13], row[14], row[15], row[16], row[17])
				tskInfo2 := TaskInfo{
					Kbn: row[1],
					GyomuKbn: row[0],
					KinoName: row[2],
					Phase: "単体設計",
					PlanStart: convert(row, 23),
					PlanEnd: convert(row, 24),
					ReViewPlanStart: convert(row, 25),
					ReViewPlanEnd: convert(row, 26),
				}

				taskInfos = append(taskInfos, tskInfo2)
				//fmt.Printf("単体設計:%s,%s,%s,%s,%s,%s,%s,%s\n", row[0], row[1], row[2], row[13], row[23], row[24], row[25], row[26])
				tskInfo3 := TaskInfo{
					Kbn: row[1],
					GyomuKbn: row[0],
					KinoName: row[2],
					Phase: "単体実施",
					PlanStart: convert(row, 32),
					PlanEnd: convert(row, 33),
					ReViewPlanStart: convert(row, 34),
					ReViewPlanEnd: convert(row, 35),
				}

				taskInfos = append(taskInfos, tskInfo3)
				//fmt.Printf("単体実施:%s,%s,%s,%s,%s,%s,%s,%s\n", row[0], row[1], row[2], row[13], row[32], row[33], safeVal(row, 34), safeVal(row, 35))
			}
		}
	}

	for _, taskInfo := range taskInfos {
		fmt.Printf("%s,%s,%s,%s,%s,%s,%s\n", taskInfo.Kbn, taskInfo.KinoName, taskInfo.Phase, taskInfo.PlanStart, taskInfo.PlanEnd, taskInfo.ReViewPlanStart, taskInfo.ReViewPlanEnd)
	}

	//checkDBMetas()
	m, _ := ReadTask2(ctx)
	b, err := json.MarshalIndent(m, "", "  ") 
	if err != nil {
		fmt.Printf("json marshal failed. %v \n", err)
	}

	fmt.Printf(" xxxxxxxxxxx %d \n xxxx %s \n", len(m), b)
	ctx.Data["TaskDatas"] = string(b)
	//checkPageCount()
	//fmt.Println("--------%s \n", convert())
	ctx.Data["PageIsMiniTask"] = true
	ctx.HTML(http.StatusOK, tplMiniTask)
    // // Get value from cell by given worksheet name and axis.
    // cell, err := f.GetCellValue("スケジュール_最新", "C5")
    // if err != nil {
    //     fmt.Println(err)
    //     return
    // }
    // fmt.Println(cell)
    // // Get all the rows in the Sheet1.
    // rows, err := f.GetRows("Sheet1")
    // if err != nil {
    //     fmt.Println(err)
    //     return
    // }
    // for _, row := range rows {
    //     for _, colCell := range row {
    //         fmt.Print(colCell, "\t")
    //     }
    //     fmt.Println()
    // }
}

func ReadTask2(ctx *context.Context) ([]TaskInfo, []string)  {
	f, err := excelize.OpenFile("D:/★FM-MAT/trunk/00_進捗管理/八丁堀内部管理用スケジュール.xlsx")
    if err != nil {
        fmt.Printf("open %s, error %v", "", err)
        return nil, nil
    }
    defer func() {
        // Close the spreadsheet.
        if err := f.Close(); err != nil {
            fmt.Printf("close %s, error %v", "", err)
        }
    }()

	rows, err := f.Rows("八丁堀内部スケジュール")
	if err != nil {
		return nil, nil
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


	xiangxi, _ := getAllFiles("D:/★FM-MAT/trunk/02_成果物/01_詳細設計")
	tantai, _ := getAllFiles("D:/★FM-MAT/trunk/02_成果物/02_単体設計")
	selfchk, _ := getAllFiles("D:/★FM-MAT/trunk/02_成果物/05_セルフチェックリスト")
	review, _ := getAllFiles("D:/★FM-MAT/trunk/02_成果物/99_レビュー記録")

	// TODO: レビュー記録表名称変更
	// for _, reviewFile := range review {
	// 	renameFile(reviewFile)
	// }

	fmt.Printf("内部設計%d, 単体%d, セルフチェックリスト%d, レビュー記録表%d\n", len(xiangxi), len(tantai), len(selfchk), len(review))
	for _, taskInfo := range taskInfos {
		
		if len(taskInfo.RealEnd) > 0 && taskInfo.RealEnd != "1899/12/30" {
			//fmt.Printf("taskInfo.RealEnd %s  Phase %s \n", taskInfo.RealEnd, taskInfo.Phase)
			if taskInfo.Phase == "内部設計" || taskInfo.Phase == "詳細設計" {
				if taskInfo.Kbn == "API" {
					checkAPIID(xiangxi, taskInfo)
					checkReviewResult(review, taskInfo)
				} else {
					checkPRID(xiangxi, taskInfo)
					checkReviewResult(review, taskInfo)
					checkIsExistSelfChk(selfchk, taskInfo)

					checkAPID(xiangxi, taskInfo)
					checkReviewResult(review, taskInfo)
					checkIsExistSelfChk(selfchk, taskInfo)
				}
			} else if taskInfo.Phase == "単体設計" {
				checkReviewResult(review, taskInfo)
				checkIsExistSelfChk(selfchk, taskInfo)
			}
		}
		//fmt.Printf("%s,%s,%s,%s,%s,%s,%s\n", taskInfo.Kbn, taskInfo.KinoName, taskInfo.Phase, taskInfo.PlanStart, taskInfo.PlanEnd, taskInfo.ReViewPlanStart, taskInfo.ReViewPlanEnd)
	}

	// 今週作成予定もの
	var weekTask1 []string  // 内部設計
	var weekTask2 []string  // 製造
	var weekTask3 []string  // 単体設計
	var weekTask4 []string  // 単体実施
	// -1：前週、0：当週、1：来週
	start, end := WeekIntervalTime(1)
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
					weekTask1 = append(weekTask1, expectFileName)
				} else {
					expectFileName := fmt.Sprintf("\t詳細設計_MAT_PR層ビジネスロジック仕様(%s)_%s.xlsx", taskInfo.KinoID, taskInfo.KinoName)
					weekTask1 = append(weekTask1, expectFileName)
					expectFileName  = fmt.Sprintf("\t詳細設計_MAT_AP層ビジネスロジック仕様(%s)_%s.xlsx", taskInfo.KinoID, taskInfo.KinoName)
					weekTask1 = append(weekTask1, expectFileName)
				}
			} else if taskInfo.Phase == "製造" {
				weekTask2 = append(weekTask2, fmt.Sprintf("\t%sのソース一式", taskInfo.KinoName))
			} else if taskInfo.Phase == "単体設計" {
				expectFileName := fmt.Sprintf("\t単体テスト仕様書兼報告書_(%s)_%s.xlsx", taskInfo.KinoID, taskInfo.KinoName)
				weekTask3 = append(weekTask3, expectFileName)
			} else if taskInfo.Phase == "単体実施" {
				expectFileName := fmt.Sprintf("\t単体テスト仕様書兼報告書_(%s)_%s_テスト結果.xlsx", taskInfo.KinoID, taskInfo.KinoName)
				weekTask4 = append(weekTask4, expectFileName)
			}
		}
		//fmt.Printf("%s,%s,%s,%s,%s,%s,%s\n", taskInfo.Kbn, taskInfo.KinoName, taskInfo.Phase, taskInfo.PlanStart, taskInfo.PlanEnd, taskInfo.ReViewPlanStart, taskInfo.ReViewPlanEnd)
	}
	var build strings.Builder
	build.WriteString("\n内部設計\n")
	build.WriteString(strings.Join(weekTask1, "\n"))
	build.WriteString("\n製造\n")
	build.WriteString(strings.Join(weekTask2, "\n"))
	build.WriteString("\n単体設計\n")
	build.WriteString(strings.Join(weekTask3, "\n"))
	build.WriteString("\n単体実施\n")
	build.WriteString(strings.Join(weekTask4, "\n"))
	fmt.Println(build.String())

	return taskInfos, []string{
		"大区分",
		"業務名",
		"機能ID",
		"機能名",
		"フェーズ",
		"予定開始日",
		"予定終了日",
		"実績開始日",
		"実績終了日",
		"担当者",
		"レ予定開始日",
		"レ予定終了日",
		"レ実績開始日",
		"レ実績終了日",
	}

}

func WeekIntervalTime(week int) (startTime, endTime time.Time) {
	now := time.Now()
	offset := int(time.Monday - now.Weekday())
	//周日做特殊判断 因为time.Monday = 0
	if offset > 0 {
	   offset = -6
	}
 
	year, month, day := now.Date()
	thisWeek := time.Date(year, month, day, 0, 0, 0, 0, time.Local)

	startTime, _ = time.Parse("2006-01-02 15:04:05", thisWeek.AddDate(0, 0, offset + 7 * week).Format("2006-01-02") + " 00:00:00")
	endTime, _ = time.Parse("2006-01-02 15:04:05", thisWeek.AddDate(0, 0, offset + 6 + 7 * week).Format("2006-01-02") + " 23:59:59")
 
	return startTime, endTime
 }

type GyomuInfo struct {
	GyomuID string
	GyomuName string
	KinoData []KinoInfo
}

type KinoInfo struct {
	Kbn string
	ID string
	Name string
}

var gyomu []GyomuInfo
// 外部設計フォルダをスキャンして外部設計のIDと機能名を取得する
func ScanFD() ([]GyomuInfo, error) {
	result := []GyomuInfo{}
	var gyomuinfo GyomuInfo

	var gyomumap = map[string]string {
		"02":"発注",
		"03":"品揃え",
		"05":"会計",
		"07":"在庫管理",
		"08":"従業員管理",
		"09":"本部連絡",
		"12":"営業管理",
		"14":"業務共通",
		"18":"CEメンテナンス",
	}

	for k, v := range gyomumap {
		gyomuinfo = GyomuInfo{
			GyomuID: k,
			GyomuName: v,
			KinoData: []KinoInfo{},
		}

		temp021, _ := getAllFiles(fmt.Sprintf("D:/★FM-MAT/trunk/01_受領資料/01_外部設計書/%s.%s/11.画面機能設計書", k, v))
		fmt.Printf("画面本数：%s.%s, %d\n", k, v, len(temp021))
		scanGamen(temp021, &gyomuinfo.KinoData)
		temp022, _ := getAllFiles(fmt.Sprintf("D:/★FM-MAT/trunk/01_受領資料/01_外部設計書/%s.%s/13.バッチ機能設計書", k, v))
		scanAPI(temp022, &gyomuinfo.KinoData)
		fmt.Printf("API本数：%s.%s, %d\n", k, v, len(temp022))

		result = append(result, gyomuinfo)
	}

	for _, r := range result {
		for _, l := range r.KinoData {
			fmt.Printf("%s,%s,%s,%s,%s\n", r.GyomuID, r.GyomuName, l.Kbn, l.ID, l.Name)
		}
	}
	return result, nil
}

func scanGamen(files []string, kindo *[]KinoInfo)  {
	reg1 := regexp.MustCompile(`外部設計_MAT_画面機能設計書\(([A-Z]{3}-[A-Z]{3}-[A-Z0-9]{4})_(.*)\) *.xlsx$`)
	for _, file := range files {
		params := reg1.FindStringSubmatch(filepath.Base(file))
		//fmt.Printf("%s -- %d\n", filepath.Base(file), len(params))
		if len(params) == 3 {
			result := KinoInfo{
				Kbn: "画面",
				ID: params[1],
				Name: params[2],
			}
			*kindo = append(*kindo, result)
		}
	}
}

func scanAPI(files []string, kindo *[]KinoInfo) {
	reg1 := regexp.MustCompile(`外部設計_ストコン_API機能設計書\(\(([a-zA-Z0-9]+)\)_\((.*)\)\) *.xlsx$`)
	for _, file := range files {
		params := reg1.FindStringSubmatch(filepath.Base(file))
		//fmt.Printf("%s -- %d\n", filepath.Base(file), len(params))
		if len(params) == 3 {
			result := KinoInfo{
				Kbn: "API",
				ID: params[1],
				Name: params[2],
			}
			*kindo = append(*kindo, result)
		}
	}
}

// ダメ
// func checkPageCount() {
// 	f, err := excelize.OpenFile("D:/★FM-MAT/trunk/02_成果物/01_詳細設計/詳細設計_MAT_AP層ビジネスロジック仕様(SRM-ACC-HK40)_期限確認.xlsx")
//     if err != nil {
//         fmt.Println(err)
//         return
//     }
//     defer func() {
//         // Close the spreadsheet.
//         if err := f.Close(); err != nil {
//             fmt.Println(err)
//         }
//     }()

// 	var count excelize.PageLayoutPaperSize
// 	for _, name := range f.GetSheetList() {
// 		// 打印pagesize？
// 		var paperSize   excelize.PageLayoutPaperSize
// 		if err := f.GetPageLayout(name, &paperSize); err != nil {
// 			fmt.Println(err)
// 		}
// 		count += paperSize
// 	}
// 	fmt.Printf(" page size %d \n", count)
// }

func checkDBMetas() {
	x, _ := getOracleDbEngine()
	tables, err := x.DBMetas()
	if err != nil {
		//return err
	}
	for _, table := range tables {
		// if _, err := x.Exec(fmt.Sprintf("ALTER TABLE `%s` ROW_FORMAT=dynamic;", table.Name)); err != nil {
		// 	return err
		// }

		// if _, err := x.Exec(fmt.Sprintf("ALTER TABLE `%s` CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;", table.Name)); err != nil {
		// 	return err
		// }
		fmt.Printf("xxx-----xxx%s\n", table.Name)

		for _, pk := range table.PrimaryKeys {
			fmt.Printf("xxxx--pk---xx%s \n", pk)
		}

		for _, col := range table.Columns() {
			fmt.Printf("xxxx--col---xx%s %b %s %s %s %s %b \n", 
				col.Name,
				 col.IsPrimaryKey, 
				 col.SQLType.Name, 
				 col.Length, 
				 col.Length2, 
				 col.Default, 
				 col.Nullable)
		}
	}

	//x.DumpAllToFile("d:/test.sql")
	//x.ImportFile(fpath string)
}

func getOracleDbEngine() (*xorm.Engine, error) {
	engine, err := xorm.NewEngine("oci8", "matuser/matuser@192.168.11.96:1521/pdb")
	if err != nil {
		//log.Fatal(err)
	}
	return engine, err
}

func convert(row []string, cols int) string {
	if cols < len(row) && len(row[cols]) > 0 {
		f, _ := strconv.ParseFloat(row[cols], 64)
		t, _ := excelize.ExcelDateToTime(f, false)
		return t.Format("2006/01/02")
	} else {
		return ""
	}
}

func convert2(row []string, cols int, isDate bool) string {
	if cols < len(row) && len(row[cols]) > 0 {
		if isDate {
			f, _ := strconv.ParseFloat(row[cols], 64)
			t, _ := excelize.ExcelDateToTime(f, false)
			return t.Format("2006/01/02")
	    } else {
			return row[cols]
		}
	} else {
		return ""
	}
}

// PR層内部設計書のチェック
func checkPRID(files []string, task TaskInfo) {
	var path string
	expectFileName := fmt.Sprintf("詳細設計_MAT_PR層ビジネスロジック仕様(%s)_%s.xlsx", task.KinoID, task.KinoName)
	//fmt.Printf("xxxxx %s \n", expectFileName)
	// ファイル存在チェック
	for _, file := range files {
		if filepath.Base(file) == expectFileName {
			path = file
			break
		}
	}

	if len(path) <= 0 {
		fmt.Printf("×　担当者:%s ファイル名：%s\n", task.PIC, expectFileName)
		return
	}

	f, err := excelize.OpenFile(path)
    if err != nil {
        fmt.Printf("open %s, error %v", path, err)
        return
    }
    defer func() {
        // Close the spreadsheet.
        if err := f.Close(); err != nil {
            fmt.Printf("close %s, error %v", path, err)
        }
    }()
}

// AP層内部設計書のチェック
func checkAPID(files []string, task TaskInfo) {
	var path string
	expectFileName := fmt.Sprintf("詳細設計_MAT_AP層ビジネスロジック仕様(%s)_%s.xlsx", task.KinoID, task.KinoName)
	// ファイル存在チェック
	for _, file := range files {
		if filepath.Base(file) == expectFileName {
			path = file
			break
		}
	}

	if len(path) <= 0 {
		fmt.Printf("×　担当者:%s ファイル名：%s\n", task.PIC, expectFileName)
		return
	}

	f, err := excelize.OpenFile(path)
    if err != nil {
        fmt.Printf("open %s, error %v", path, err)
        return
    }
    defer func() {
        // Close the spreadsheet.
        if err := f.Close(); err != nil {
            fmt.Printf("close %s, error %v", path, err)
        }
    }()

}

// API内部設計書のチェック
func checkAPIID(files []string, task TaskInfo) {
	var path string
	expectFileName := fmt.Sprintf("詳細設計_MAT_API_SQL仕様((%s)_(%s)).xlsx", task.KinoID, task.KinoName)
	// ファイル存在チェック
	for _, file := range files {
		if filepath.Base(file) == expectFileName {
			path = file
			break
		}
	}

	if len(path) <= 0 {
		fmt.Printf("×　担当者:%s ファイル名：%s\n", task.PIC, expectFileName)
		return
	}

	f, err := excelize.OpenFile(path)
    if err != nil {
        fmt.Printf("open %s, error %v", path, err)
        return
    }
    defer func() {
        // Close the spreadsheet.
        if err := f.Close(); err != nil {
            fmt.Printf("close %s, error %v", path, err)
        }
    }()
}

// レビュー記録表のチェック
func checkReviewResult(files []string, task TaskInfo) {
	var path string

	if len(task.ReViewRealEnd) == 0 || task.ReViewRealEnd != "1899/12/30" {
		return
	}

	var pattern string
	if task.Phase =="詳細設計" || task.Phase =="内部設計" {
		if task.Kbn == "API" {
			pattern = fmt.Sprintf("レビュー記録表(内部設計)_20220000_API_(%s_%s).xlsm", task.KinoID, task.KinoName)
			kinoName := strings.Replace(task.KinoName, "(", `\(`, -1)
			kinoName = strings.Replace(kinoName, ")", `\)`, -1)
			reg1 := regexp.MustCompile(fmt.Sprintf(`レビュー記録表\(内部設計\)_[0-9]{8}_API_\(%s_%s\).xlsm`, task.KinoID, kinoName))
			// ファイル存在チェック
			for _, file := range files {
				if reg1.MatchString(filepath.Base(file)) {
					path = file
					break
				}
			}

			if len(path) <= 0 {
				fmt.Printf("×　担当者:%s レビュー記録表：%s\n", task.PIC, pattern)
				return
			}
		} else {
			pattern = fmt.Sprintf("レビュー記録表(内部設計)_20220000_PR_(%s_%s).xlsm", task.KinoID, task.KinoName)
			kinoName := strings.Replace(task.KinoName, "(", `\(`, -1)
			kinoName = strings.Replace(kinoName, ")", `\)`, -1)
			reg1 := regexp.MustCompile(fmt.Sprintf(`レビュー記録表\(内部設計\)_[0-9]{8}_PR_\(%s_%s\).xlsm`, task.KinoID, kinoName))
			// ファイル存在チェック
			for _, file := range files {
				if reg1.MatchString(filepath.Base(file)) {
					path = file
					break
				}
			}

			if len(path) <= 0 {
				fmt.Printf("×　担当者:%s レビュー記録表：%s\n", task.PIC, pattern)
				return
			}

			pattern = fmt.Sprintf("レビュー記録表(内部設計)_20220000_AP_(%s_%s).xlsm", task.KinoID, task.KinoName)
			kinoName = strings.Replace(task.KinoName, "(", `\(`, -1)
			kinoName = strings.Replace(kinoName, ")", `\)`, -1)
			reg1 = regexp.MustCompile(fmt.Sprintf(`レビュー記録表\(内部設計\)_[0-9]{8}_AP_\(%s_%s\).xlsm`, task.KinoID, kinoName))
			// ファイル存在チェック
			for _, file := range files {
				if reg1.MatchString(filepath.Base(file)) {
					path = file
					break
				}
			}

			if len(path) <= 0 {
				fmt.Printf("×　担当者:%s レビュー記録表：%s\n", task.PIC, pattern)
				return
			}
		}
	}

	if len(path) <= 0 {
		return
	}

	f, err := excelize.OpenFile(path)
    if err != nil {
		fmt.Printf("open %s, error %v", path, err)
        return
    }
    defer func() {
        // Close the spreadsheet.
        if err := f.Close(); err != nil {
            fmt.Printf("close %s, error %v", path, err)
        }
    }()
}

// セルフチェックリストの有無チェック
func checkIsExistSelfChk(files []string, task TaskInfo) {

}

// 全てのファイルを取得する
func getAllFiles(pathname string) ([]string, error) {
	result := []string{}

	fis, err := ioutil.ReadDir(pathname)
	if err != nil {
		fmt.Printf("read directory failed. pathname=%v, err=%v", pathname, err)
		return result, err
	}

	for _, fi := range fis {
		fullname := pathname + "/" + fi.Name()
		if fi.IsDir() {
			if fi.Name() != "bak" && fi.Name() != "00_PR・AP・単体詳細レビューフォーマット" {
				temp, err := getAllFiles(fullname)
				if err != nil {
					fmt.Printf("read directory failed. pathname=%v, err=%v", pathname, err)
					return result, err
				}
				result = append(result, temp...)
			}
		} else {
			ext := filepath.Ext(fi.Name())
			//fmt.Println(ext)
			if ext == ".xlsx" || ext == ".xls" || ext == ".xlsm" {
				result = append(result, fullname)
			}
		}
	}
	return result, nil
}

// oldpattern:レビュー記録表(内部設計)_20220607_AP_(SRM-ASS-TR22)_ゴンドラ選択ポップアップ.xlsm
// newpattern:レビュー記録表(内部設計)_yyyymmdd_AP_(SRM-ACC-HP30_雑誌返品入力).xlsm
func renameFile(filename string) {
	regex := regexp.MustCompile(`(.*)_\(([A-Z0-9]{3}-[A-Z0-9]{3}-[A-Z0-9]{4})\)_(.*).xlsm`)
	s := regex.ReplaceAllString(filename, "${1}_(${2}_${3}).xlsm")
	//fmt.Printf("xxxx %s %s \n", filename, s)
	err := os.Rename(filename,  s)
	if err != nil {
		fmt.Printf(`Renameに失敗: %s\n`, filename)
	}
}

// 文字列が数値かを判断する
func isNum(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}
