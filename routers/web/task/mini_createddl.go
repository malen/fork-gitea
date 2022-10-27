package task

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"

	"code.gitea.io/gitea/modules/base"
	"code.gitea.io/gitea/modules/context"
	"github.com/xuri/excelize/v2"
)

const (
	tplMiniCreateDDL base.TplName = "task/create_ddl"
)

type GLimit struct {
	n int
	c chan struct{}
}

func New(n int) *GLimit {
	return &GLimit{
		n: n,
		c: make(chan struct{}, n),
	}
}

func (g *GLimit) Run(f func()) {
	g.c <- struct{}{}
	go func() {
		f()
		<-g.c
	}()
}

type TableInfo struct {
	TblID     string
	TblName   string
	Columns   []TableColumnInfo
	PkColumns []string
}

type TableColumnInfo struct {
	ColID    string
	ColName  string
	ColType  string
	ColLen   string
	Nullable string
	Default  string
}

func CreateDDL(ctx *context.Context) {

	targetTbls := [...]string{
		"TACT_KTM_NHN_DNP",
		"TACT_KTM_NHN_DNP_SHN_MS",
		"TACT_KTM_NHN_DNP_TRAN",
		"TCMC_DATE_KANRI",
		"TMMM_SHOHIN_TENPO_ZAIKO",
		"VODO_HTTNK_3_R3KN_LWEK",
		"TMMM_SHOHIN",
		"TMMM_SHOHIN_BIN",
		"VW_CAN_DTE_ORD",
		"VW_CAN_DTE_REV_ORD",
		"TASH_SHN_SNZR_RRK",
		"TMMM_HATCHU_CHUDAN_TAISHO",
		"TMMM_ZAIKO_BUMPAI_TAISHO",
		"TODC_HATCHU_KANRI",
	}

	sqlExcel, _ := getAllFiles("D:/★FM-MAT/trunk/01_受領資料/04_テーブル定義/SC/テーブルレイアウト(API用)")
	reg1 := regexp.MustCompile(`.*：(.*)\(([A-Z0-9_]*)\)`)

	var tableInfos []TableInfo
	var tab string
	fmt.Printf("sql excel count: %d \n", len(sqlExcel))

	// 多协程
	var wg = sync.WaitGroup{}
	g := New(10)

	for _, file := range sqlExcel {

		wg.Add(1)
		path := file
		goFunc := func() {

			f, err := excelize.OpenFile(path)
			if err != nil {
				fmt.Printf("open %s, error %v", "", err)
			}

			defer func() {
				// Close the spreadsheet.
				if err := f.Close(); err != nil {
					fmt.Printf("close %s, error %v", path, err)
				}
				wg.Done()
			}()

			cell, err := f.GetCellValue("ファイルレイアウト(表)", "A7")
			if err != nil {
				//fmt.Printf("xxxxx %s  %v \n", file, err)
				return
			}

			// テーブル名を取得する
			params := reg1.FindStringSubmatch(cell)
			//fmt.Printf("%s -- %d\n", filepath.Base(file), len(params))
			if len(params) == 3 {

				//fmt.Printf("table: %s, %s \n", params[1], params[2])
			} else {
				return
			}

			tab = ""
			for _, target := range targetTbls {

				if target == params[2] {
					tab = params[2]
					break
				}
			}

			if len(tab) == 0 {
				return
			}

			tableInfo := TableInfo{
				TblID:     params[2],
				TblName:   params[1],
				Columns:   []TableColumnInfo{},
				PkColumns: []string{},
			}

			rows, _ := f.Rows("ファイルレイアウト(表)")
			b := false
			for rows.Next() {
				row, _ := rows.Columns()

				if row[0] == "1" {
					b = true
				}

				if b && len(row[0]) > 0 {

					// pk
					if len(row[1]) > 0 {
						tableInfo.PkColumns = append(tableInfo.PkColumns, row[7])
					}

					// columns
					col := TableColumnInfo{
						ColID:    row[7],
						ColName:  row[6],
						ColType:  row[8],
						ColLen:   row[9],
						Nullable: row[10],
						Default:  row[11],
					}
					tableInfo.Columns = append(tableInfo.Columns, col)
					//fmt.Printf("xxxx %v    %v \n", row, col.ColID)
				}
			}
			tableInfos = append(tableInfos, tableInfo)

		}

		g.Run(goFunc)
		//fmt.Printf("table: %s, %s \n", params[1], params[2])
	}
	wg.Wait()

	var build strings.Builder
	//fmt.Printf("------------------%d \n", len(tableInfos))
	for _, tableInfo := range tableInfos {
		build.Reset()
		build.WriteString(fmt.Sprintf("--DROP TABLE %s CASCADE CONSTRAINT PURGE;\n", tableInfo.TblID))
		build.WriteString(fmt.Sprintf("CREATE TABLE %s (\n", tableInfo.TblID))

		colData := []string{}
		for _, col := range tableInfo.Columns {
			if len(col.ColLen) > 0 {
				colData = append(colData, strings.TrimRight(fmt.Sprintf("    %s %s(%s) %s %s", col.ColID, col.ColType, col.ColLen, col.Nullable, col.Default), " "))
			} else {
				//build.WriteString(fmt.Sprintf("    %s %s %s %s,\n", col.ColID, col.ColType, col.Nullable, col.Default))
				colData = append(colData, strings.TrimRight(fmt.Sprintf("    %s %s %s %s", col.ColID, col.ColType, col.Nullable, col.Default), " "))
			}
			//fmt.Printf("table: %s, %s col cout %d\n", tableInfo.TblID, tableInfo.TblName, len(tableInfo.Columns))
		}
		build.WriteString(strings.Join(colData, ",\n"))
		build.WriteString("\n);\n")

		if len(tableInfo.PkColumns) > 0 {
			build.WriteString(fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT PK_%s PRIMARY KEY (%s);\n", tableInfo.TblID, tableInfo.TblID, strings.Join(tableInfo.PkColumns, ",")))
		}

		build.WriteString(fmt.Sprintf("COMMENT ON TABLE %s IS '%s';\n", tableInfo.TblID, tableInfo.TblName))
		for _, col := range tableInfo.Columns {
			build.WriteString(fmt.Sprintf("COMMENT ON COLUMN %s.%s IS '%s';\n", tableInfo.TblID, col.ColID, col.ColName))
		}
		s3 := build.String()

		fmt.Println(s3)
	}

	ctx.Data["PageIsMiniTask"] = true
	ctx.HTML(http.StatusOK, tplMiniCreateDDL)
}

func CreateDDLByFilePath(ctx *context.Context) {
	if len(ctx.FormString("file")) == 0 {
		ctx.Data["PageIsMiniTask"] = true
		ctx.HTML(http.StatusOK, tplMiniCreateDDL)
		return
	}

	sql, err := create(ctx.FormString("file"))
	if err != nil {
		// ctx.Data["SqlDDL"] = err.
	} else {
		ctx.Data["SqlDDL"] = sql
	}

	ctx.Data["FilePath"] = ctx.FormString("file")
	ctx.Data["PageIsMiniTask"] = true
	ctx.HTML(http.StatusOK, tplMiniCreateDDL)
	return
}

func create(path string) (string, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		errMsg := fmt.Sprintf("open %s, error %v", "", err)
		return "", errors.New(errMsg)
	}

	defer func() {
		if err := f.Close(); err != nil {
			//errMsg := fmt.Sprintf("close %s, error %v", path, err)
			//return "", errors.New(errMsg)
		}
	}()

	cell, err := f.GetCellValue("ファイルレイアウト(表)", "A7")
	if err != nil {
		//fmt.Printf("xxxxx %s  %v \n", file, err)
		errMsg := fmt.Sprintf("read %s  %v \n", path, err)
		return "", errors.New(errMsg)
	}

	reg1 := regexp.MustCompile(`.*：(.*)\(([A-Z0-9_]*)\)`)
	// テーブル名を取得する
	params := reg1.FindStringSubmatch(cell)
	//fmt.Printf("%s -- %d\n", filepath.Base(file), len(params))
	if len(params) == 3 {

		//fmt.Printf("table: %s, %s \n", params[1], params[2])
	} else {
		return "", errors.New("テーブル名が正しく取得できません。")
	}

	// tab = ""
	// for _, target := range targetTbls {

	// 	if target == params[2] {
	// 		tab = params[2]
	// 		break
	// 	}
	// }

	// if len(tab) == 0 {
	// 	return
	// }

	tableInfo := TableInfo{
		TblID:     params[2],
		TblName:   params[1],
		Columns:   []TableColumnInfo{},
		PkColumns: []string{},
	}

	rows, _ := f.Rows("ファイルレイアウト(表)")
	b := false
	for rows.Next() {
		row, _ := rows.Columns()

		if row[0] == "1" {
			b = true
		}

		if b && len(row[0]) > 0 {

			// pk
			if len(row[1]) > 0 {
				tableInfo.PkColumns = append(tableInfo.PkColumns, row[7])
			}

			// columns
			col := TableColumnInfo{
				ColID:    row[7],
				ColName:  row[6],
				ColType:  row[8],
				ColLen:   row[9],
				Nullable: row[10],
				Default:  row[11],
			}
			tableInfo.Columns = append(tableInfo.Columns, col)
			//fmt.Printf("xxxx %v    %v \n", row, col.ColID)
		}
	}

	var build strings.Builder
	build.Reset()
	build.WriteString(fmt.Sprintf("--DROP TABLE %s CASCADE CONSTRAINT PURGE;\n", tableInfo.TblID))
	build.WriteString(fmt.Sprintf("CREATE TABLE %s (\n", tableInfo.TblID))

	colData := []string{}
	for _, col := range tableInfo.Columns {
		if len(col.ColLen) > 0 {
			colData = append(colData, strings.TrimRight(fmt.Sprintf("    %s %s(%s) %s %s", col.ColID, col.ColType, col.ColLen, formatDefault(col.Default), col.Nullable), " "))
		} else {
			//build.WriteString(fmt.Sprintf("    %s %s %s %s,\n", col.ColID, col.ColType, col.Nullable, col.Default))
			colData = append(colData, strings.TrimRight(fmt.Sprintf("    %s %s %s %s", col.ColID, col.ColType, formatDefault(col.Default), col.Nullable), " "))
		}
		//fmt.Printf("table: %s, %s col cout %d\n", tableInfo.TblID, tableInfo.TblName, len(tableInfo.Columns))
	}
	build.WriteString(strings.Join(colData, ",\n"))
	build.WriteString("\n);\n")

	if len(tableInfo.PkColumns) > 0 {
		build.WriteString(fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT PK_%s PRIMARY KEY (%s);\n", tableInfo.TblID, tableInfo.TblID, strings.Join(tableInfo.PkColumns, ",")))
	}

	build.WriteString(fmt.Sprintf("COMMENT ON TABLE %s IS '%s';\n", tableInfo.TblID, tableInfo.TblName))
	for _, col := range tableInfo.Columns {
		build.WriteString(fmt.Sprintf("COMMENT ON COLUMN %s.%s IS '%s';\n", tableInfo.TblID, col.ColID, col.ColName))
	}
	s3 := build.String()

	writeToFile(s3, tableInfo.TblID, tableInfo.TblName)
	return s3, nil
}

func writeToFile(sql string, tblID string, tblName string) {
	fileName := fmt.Sprintf(`D:/★FM-MAT/trunk/01_受領資料/04_テーブル定義/SC/テーブルレイアウト(API用)/DDL/%s_%s.sql`, tblID, tblName)

	// os.OpenFile の第二引数に指定可能なオプション
	// os.O_RDONLY, syscall.O_RDONLY
	// 読み込み専用
	// os.O_WRONLY, syscall.O_WRONLY
	// 書き込み専用
	// os.O_RDWR, syscall.O_RDWR
	// 読み書き両方
	// os.O_APPEND, syscall.O_APPEND
	// 上書きではなく追記する
	// os.O_CREATE, syscall.O_CREAT
	// ファイルを作成する
	// os.O_EXCL, syscall.O_EXCL
	// O_CREATEと一緒に指定する。ファイルが存在した場合にエラーになる
	// os.O_SYNC, syscall.O_SYNC
	// 同期書き込み(同期I/O)。書き込みが完了するまで次の処理はブロックされる
	// os.O_TRUNC, syscall.O_TRUNC

	// ファイルを開くときに切り詰める
	f, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0664) // ファイルを読み取りで開く
	if err != nil {
		fmt.Printf(`ファイル作成に失敗しました %s \n`, fileName)
	}
	defer func() {
		if err := f.Close(); err != nil {
			//return err
		}
	}()

	f.WriteString(sql)
}

func formatDefault(defvalue string) string {
	if len(defvalue) > 0 {
		return fmt.Sprintf("DEFAULT %s", defvalue)
	} else {
		return ""
	}
}
