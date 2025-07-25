package tasks

import (
	"gantt/internal/entity/postgresql/db/projects"
	"gantt/internal/entity/postgresql/db/s3_files"
	"gantt/internal/entity/postgresql/db/task_resources"
	"gantt/internal/entity/postgresql/db/users"
	"gantt/internal/interactor/models/special"
	model "gantt/internal/interactor/models/tasks"
	"time"
)

// Table struct is tasks database table struct
type Table struct {
	// 表ID
	TaskUUID string `gorm:"<-:create;column:task_uuid;type:uuid;not null;primaryKey;" json:"task_uuid"`
	// 前端編號 (非表ID)
	TaskOldID int64 `gorm:"->;column:task_old_id;type:serial;" json:"task_old_id"`
	// 前端編號 (非表ID)
	TaskID string `gorm:"column:task_id;type:text;" json:"task_id"`
	// 任務名稱
	TaskName string `gorm:"column:task_name;type:text;" json:"task_name"`
	// 起始日期
	StartDate *time.Time `gorm:"column:start_date;type:timestamp;" json:"start_date"`
	// 結束日期
	EndDate *time.Time `gorm:"column:end_date;type:timestamp;" json:"end_date"`
	// 基準線起始日期
	BaselineStartDate *time.Time `gorm:"column:baseline_start_date;type:timestamp;" json:"baseline_start_date"`
	// 基準線結束日期
	BaselineEndDate *time.Time `gorm:"column:baseline_end_date;type:timestamp;" json:"baseline_end_date"`
	// 基準線工作天
	BaselineDuration float64 `gorm:"column:baseline_duration;type:numeric;" json:"baseline_duration"`
	// 期間
	Duration float64 `gorm:"column:duration;type:numeric;" json:"duration"`
	// 完成百分比
	Progress int64 `gorm:"column:progress;type:int;" json:"progress"`
	// 花費時間
	Cost int64 `gorm:"column:cost;type:int;" json:"cost"`
	// 前任
	Predecessor string `gorm:"column:predecessor;type:text;" json:"predecessor"`
	// 1.1.2、1.2、1.2.1
	OutlineNumber string `gorm:"column:outline_number;type:text;" json:"outline_number"`
	// 未知
	Assignments string `gorm:"column:assignments;type:text;" json:"assignments"`
	// 紀錄標的顏色
	TaskColor string `gorm:"column:task_color;type:text;" json:"task_color"`
	// 預留：外部連結
	WebLink string `gorm:"column:web_link;type:text;" json:"web_link"`
	// 是否為任務
	IsSubTask bool `gorm:"column:is_subtask;type:boolean;default:false" json:"is_subtask"`
	// 專案UUID
	ProjectUUID *string `gorm:"column:project_uuid;type:uuid;" json:"project_uuid"`
	// projects data
	Projects projects.Table `gorm:"foreignKey:ProjectUUID;references:ProjectUUID" json:"projects,omitempty"`
	// 任務分段(陣列的字串型態)
	Segment string `gorm:"column:segment;type:text;" json:"segment"`
	// 任務標示(陣列的字串型態)
	Indicator string `gorm:"column:indicator;type:text;" json:"indicator"`
	// 備註
	Notes string `gorm:"column:notes;type:text;" json:"notes"`
	// task_resources data
	TaskResources []task_resources.Table `gorm:"foreignKey:TaskUUID;" json:"resources,omitempty"`
	// s3_files data
	S3Files []s3_files.Table `gorm:"foreignKey:SourceUUID;references:TaskUUID" json:"files,omitempty"`
	// create_users data
	CreatedByUsers users.Table `gorm:"foreignKey:ID;references:CreatedBy" json:"created_by_users,omitempty"`
	// update_users data
	UpdatedByUsers users.Table `gorm:"foreignKey:ID;references:UpdatedBy" json:"updated_by_users,omitempty"`
	// 引入後端專用
	special.Table
}

// Base struct is corresponding to tasks table structure file
type Base struct {
	// 表ID
	TaskUUID *string `json:"task_uuid,omitempty"`
	// 前端編號 (非表ID)
	TaskOldID *int64 `json:"task_old_id,omitempty"`
	// 前端編號 (非表ID)
	TaskID *string `json:"task_id,omitempty"`
	// 任務名稱
	TaskName *string `json:"task_name,omitempty"`
	// 起始日期
	StartDate *time.Time `json:"start_date,omitempty"`
	// 結束日期
	EndDate *time.Time `json:"end_date,omitempty"`
	// 基準線起始日期
	BaselineStartDate *time.Time `json:"baseline_start_date,omitempty"`
	// 基準線結束日期
	BaselineEndDate *time.Time `json:"baseline_end_date,omitempty"`
	// 基準線工作天
	BaselineDuration *float64 `json:"baseline_duration,omitempty"`
	// 期間
	Duration *float64 `json:"duration,omitempty"`
	// 完成百分比
	Progress *int64 `json:"progress,omitempty"`
	// 花費時間
	Cost *int64 `json:"cost,omitempty"`
	// 前任
	Predecessor *string `json:"predecessor,omitempty"`
	// 1.1.2、1.2、1.2.1
	OutlineNumber *string `json:"outline_number,omitempty"`
	// 未知
	Assignments *string `json:"assignments,omitempty"`
	// 紀錄標的顏色
	TaskColor *string `json:"task_color,omitempty"`
	// 預留：外部連結
	WebLink *string `json:"web_link,omitempty"`
	// 是否為任務
	IsSubTask *bool `json:"is_subtask,omitempty"`
	// 專案UUID
	ProjectUUID *string `json:"project_uuid,omitempty"`
	// projects data
	Projects projects.Table `json:"projects,omitempty"`
	// 任務分段(陣列的字串型態)
	Segment *string `json:"segment,omitempty"`
	// 任務標示(陣列的字串型態)
	Indicator *string `json:"indicator,omitempty"`
	// 備註
	Notes *string `json:"notes,omitempty"`
	// task_resources data
	TaskResources []task_resources.Base `json:"resources,omitempty"`
	// s3_files data
	S3Files []s3_files.Base `json:"files,omitempty"`
	// create_users data
	CreatedByUsers users.Base `json:"created_by_users,omitempty"`
	// update_users data
	UpdatedByUsers users.Base `json:"updated_by_users,omitempty"`
	// 引入後端專用
	special.Base
	// 搜尋欄位
	model.Filter `json:"filter"`
	// 後端刪除任務及更新專案start_date及end_date用
	DeletedTaskUUIDs []*string `json:"task_uuids,omitempty"`
}

func (t *Table) TableName() string {
	return "tasks"
}
