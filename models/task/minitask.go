package task

import (
	"code.gitea.io/gitea/models/db"
	"code.gitea.io/gitea/modules/timeutil"
)

type UserType int //revive:disable-line:exported

type MiniTask struct {
	ID        int64  `xorm:"pk autoincr"`
	IDDisp	  int64  `xorm:"NOT NULL"`     // 表示用序号
    Kbn		  int `xorm:"NOT NULL"`        // 大区分 1：画面 2：API
	Status    int `xorm:"NOT NULL"`        // 未着手、作成中、作成済、製造中、製造済、単体実施中、単体実施済、レビュー依頼、指摘あり、指摘対応済、納品可、納品済、納品返品、納品再送、仕変あり、仕変対応済、QA中
	GyomuKbn  int `xorm:"NOT NULL"`        // 業務区分
	KinoID    string `xorm:"VARCHAR(12)"`  // 機能ID
	KinoName  string `xorm:"VARCHAR(64)"`  // 機能名
	Phase     string `xorm:"VARCHAR(12)"`  // 内部設計、製造、単体設計、単体実施
	PIC       int                          // 担当者
	PlanStart timeutil.TimeStamp           // 予定開始日
	PlanEnd   timeutil.TimeStamp           // 予定終了日
	RealStart timeutil.TimeStamp           // 実績開始日
	RealEnd   timeutil.TimeStamp           // 実績終了日
	CompRate  int `DEFAULT 0"`             // 進捗率
	ManHours  int                          // 工数
	PagePRCnt int                          // PR層のページ数
	PageAPCnt int                          // AP層のページ数
	PageAPICnt int                         // APIのページ数
	ReViewPlanStart timeutil.TimeStamp     // レビュー予定開始日
	ReViewPlanEnd   timeutil.TimeStamp     // レビュー予定終了日
	ReViewRealStart timeutil.TimeStamp     // レビュー実績開始日
	ReViewRealEnd   timeutil.TimeStamp     // レビュー実績終了日
}

type MiniStatus struct {
	ID         int64  `xorm:"pk autoincr"`
	StatusName string `xorm:"VARCHAR(64)"`
}

type MiniKbn struct {
	ID         int64  `xorm:"pk autoincr"`
	KbnName    string `xorm:"VARCHAR(64)"`
}

type MiniPhase struct {
	ID         int64  `xorm:"pk autoincr"`
	PhaseName  string `xorm:"VARCHAR(64)"`
}

func init() {
	db.RegisterModel(new(MiniTask))
	db.RegisterModel(new(MiniStatus))
	db.RegisterModel(new(MiniKbn))
	db.RegisterModel(new(MiniPhase))
}

func FindTasks() {
	// if _, err = db.Exec(ctx, "UPDATE `user` SET num_followers = num_followers + 1 WHERE id = ?", followID); err != nil {
	// 	return err
	// }
}

func GetAllTasks() ([]*MiniTask, error) {
	tasks := make([]*MiniTask, 0)
	return tasks, db.GetEngine(db.DefaultContext).OrderBy("id").Find(&tasks)
}