package tasks

import (
	"encoding/csv"
	"gantt/internal/interactor/models/event_marks"
	"gantt/internal/interactor/models/page"
	"gantt/internal/interactor/models/resources"
	"gantt/internal/interactor/models/s3_files"
	"gantt/internal/interactor/models/section"
	"time"
)

// Create struct is used to create achieves
type Create struct {
	// 父層級UUID
	ParentUUID *string `json:"parent_uuid,omitempty" binding:"omitempty,uuid4" validate:"omitempty,uuid4"`
	// 前端編號 (非表ID)
	TaskID string `json:"task_id,omitempty" binding:"required" validate:"required"`
	// 任務名稱
	TaskName string `json:"task_name,omitempty"`
	// 起始日期
	StartDate *time.Time `json:"start_date,omitempty"`
	// 結束日期
	EndDate *time.Time `json:"end_date,omitempty"`
	// 基準線起始日期
	BaselineStartDate *time.Time `json:"baseline_start_date,omitempty"`
	// 基準線結束日期
	BaselineEndDate *time.Time `json:"baseline_end_date,omitempty"`
	// 基準線工作天
	BaselineDuration float64 `json:"baseline_duration,omitempty"`
	// 期間
	Duration float64 `json:"duration,omitempty" binding:"omitempty,gte=0" validate:"omitempty,gte=0"`
	// 完成百分比
	Progress int64 `json:"progress,omitempty" binding:"omitempty,gte=0" validate:"omitempty,gte=0"`
	// 花費時間
	Cost int64 `json:"cost,omitempty" binding:"omitempty,gte=0" validate:"omitempty,gte=0"`
	// 人力資源
	Resources []*resources.TaskSingle `json:"resources,omitempty"`
	// 前任
	Predecessor string `json:"predecessor,omitempty"`
	// 1.1.2、1.2、1.2.1
	OutlineNumber string `json:"outline_number,omitempty"`
	// 未知
	Assignments string `json:"assignments,omitempty"`
	// 紀錄標的顏色
	TaskColor string `json:"task_color,omitempty"`
	// 預留：外部連結
	WebLink string `json:"web_link,omitempty"`
	// 是否為任務
	IsSubTask bool `json:"is_subtask,omitempty"`
	// 專案UUID
	ProjectUUID string `json:"project_uuid,omitempty" binding:"required,uuid4" validate:"required,uuid4"`
	// 子任務
	Subtask []*Create `json:"subtasks,omitempty"`
	// 任務分段
	Segments []*Segments `json:"segments,omitempty"`
	// 任務分段(陣列的字串型態)
	Segment string `json:"segment,omitempty" swaggerignore:"true"`
	// 任務標示
	Indicators []*Indicators `json:"indicators,omitempty"`
	// 任務標示(陣列的字串型態)
	Indicator string `json:"indicator,omitempty" swaggerignore:"true"`
	// 備註
	Notes string `json:"notes,omitempty"`
	// 創建者
	CreatedBy string `json:"created_by,omitempty" binding:"omitempty,uuid4" validate:"omitempty,uuid4" swaggerignore:"true"`
	// 資源UUID
	ResUUID *string `json:"res_uuid,omitempty" swaggerignore:"true"`
	// 角色
	Role *string `json:"role,omitempty" swaggerignore:"true"`
}

// Field is structure file for search
type Field struct {
	// 表ID
	TaskUUID string `json:"task_uuid,omitempty" binding:"omitempty,uuid4" validate:"omitempty,uuid4" swaggerignore:"true"`
	// 前端編號 (非表ID)
	TaskOldID *int64 `json:"task_old_id,omitempty" form:"task_old_id"`
	// 前端編號 (非表ID)
	TaskID *string `json:"task_id,omitempty"`
	// 任務名稱
	TaskName *string `json:"task_name,omitempty" form:"task_name"`
	// 起始日期
	StartDate *time.Time `json:"start_date,omitempty" form:"start_date"`
	// 結束日期
	EndDate *time.Time `json:"end_date,omitempty" form:"end_date"`
	// 基準線起始日期
	BaselineStartDate *time.Time `json:"baseline_start_date,omitempty" form:"baseline_start_date"`
	// 基準線結束日期
	BaselineEndDate *time.Time `json:"baseline_end_date,omitempty" form:"baseline_end_date"`
	// 基準線工作天
	BaselineDuration *float64 `json:"baseline_duration,omitempty" form:"baseline_duration"`
	// 期間
	Duration *float64 `json:"duration,omitempty" form:"duration"`
	// 完成百分比
	Progress *int64 `json:"progress,omitempty" form:"progress"`
	// 花費時間
	Cost *int64 `json:"cost,omitempty" form:"cost"`
	// 前任
	Predecessor *string `json:"predecessor,omitempty" form:"predecessor"`
	// 1.1.2、1.2、1.2.1
	OutlineNumber *string `json:"outline_number,omitempty" form:"outline_number"`
	// 未知
	Assignments *string `json:"assignments,omitempty" form:"assignments"`
	// 紀錄標的顏色
	TaskColor *string `json:"task_color,omitempty" form:"task_color"`
	// 預留：外部連結
	WebLink *string `json:"web_link,omitempty" form:"web_link"`
	// 是否為任務
	IsSubTask *bool `json:"is_subtask,omitempty" form:"is_subtask"`
	// 專案UUID
	ProjectUUID *string `json:"project_uuid,omitempty" form:"project_uuid"`
	// 多筆刪除任務及更新專案start_date及end_date用
	DeletedTaskUUIDs []*string `json:"task_uuids,omitempty" form:"task_uuids"`
	// 搜尋欄位
	Filter `json:"filter"`
}

// Fields is the searched structure file (including pagination)
type Fields struct {
	// 搜尋結構檔
	Field
	// 分頁搜尋結構檔
	page.Pagination
}

// Filter struct is used to store the search field
type Filter struct {
	// 是否有里程被
	FilterMilestone bool `json:"is_milestone,omitempty"`
}

// List is multiple return structure files
type List struct {
	// 多筆
	Tasks []*Single `json:"tasks"`
	// event_marks
	EventMarks []*event_marks.Single `json:"event_marks"`
	// 專案狀態
	ProjectStatus string `json:"project_status,omitempty"`
	// 是否可編輯專案任務
	IsEditable bool `json:"is_editable,omitempty"`
	// 專案開始日期
	ProjectStartDate *time.Time `json:"project_start_date,omitempty"`
	// 分頁返回結構檔
	page.Total
}

// Single return structure file
type Single struct {
	// 表ID
	TaskUUID string `json:"task_uuid,omitempty"`
	// 前端編號 (非表ID)
	TaskOldID int64 `json:"task_old_id,omitempty"`
	// 前端編號 (非表ID)
	TaskID string `json:"task_id,omitempty"`
	// 任務名稱
	TaskName string `json:"task_name,omitempty"`
	// 起始日期
	StartDate *time.Time `json:"start_date,omitempty"`
	// 結束日期
	EndDate *time.Time `json:"end_date,omitempty"`
	// 基準線起始日期
	BaselineStartDate *time.Time `json:"baseline_start_date,omitempty"`
	// 基準線結束日期
	BaselineEndDate *time.Time `json:"baseline_end_date,omitempty"`
	// 基準線工作天
	BaselineDuration float64 `json:"baseline_duration,omitempty"`
	// 期間
	Duration float64 `json:"duration,omitempty"`
	// 完成百分比
	Progress int64 `json:"progress,omitempty"`
	// 花費時間
	Cost int64 `json:"cost,omitempty"`
	// 前任
	Predecessor string `json:"predecessor,omitempty"`
	// 1.1.2、1.2、1.2.1
	OutlineNumber string `json:"outline_number,omitempty"`
	// 未知
	Assignments string `json:"assignments,omitempty"`
	// 紀錄標的顏色
	TaskColor string `json:"task_color,omitempty"`
	// 預留：外部連結
	WebLink string `json:"web_link,omitempty"`
	// 是否為任務
	IsSubTask bool `json:"is_subtask,omitempty"`
	// 專案UUID
	ProjectUUID string `json:"project_uuid,omitempty"`
	// 備註
	Notes string `json:"notes,omitempty"`
	// 任務標示名稱
	IndicatorsName string `json:"indicatorsName,omitempty"`
	// 任務標示工具提示
	IndicatorsToolTip string `json:"indicatorsTooltip,omitempty"`
	// 任務標示IconClass
	IndicatorsIconClass string `json:"indicatorsClass,omitempty"`
	// 是否可編輯或刪除任務
	IsEditable bool `json:"is_editable,omitempty"`
	// 創建者
	CreatedBy string `json:"created_by,omitempty"`
	// 更新者
	UpdatedBy string `json:"updated_by,omitempty"`
	// 時間戳記
	section.TimeAt
	// 人力資源(陣列)
	Resources []resources.TaskSingle `json:"resources,omitempty"`
	// 任務分段(陣列)
	Segments []Segments `json:"segments,omitempty"`
	// 任務標示
	Indicators []Indicators `json:"indicators,omitempty"`
	// 子任務
	Subtask []*Single `json:"subtasks,omitempty"`
	// 附件檔案
	Files []*s3_files.Single `json:"files,omitempty"`
}

// Update struct is used to update achieves
type Update struct {
	// 表ID
	TaskUUID string `json:"task_uuid,omitempty"  binding:"required,uuid4" validate:"required,uuid4" swaggerignore:"true"`
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
	Duration *float64 `json:"duration,omitempty" binding:"omitempty,gte=0" validate:"omitempty,gte=0"`
	// 完成百分比
	Progress *int64 `json:"progress,omitempty" binding:"omitempty,gte=0" validate:"omitempty,gte=0"`
	// 花費時間
	Cost *int64 `json:"cost,omitempty" binding:"omitempty,gte=0" validate:"omitempty,gte=0"`
	// 人力資源
	Resources []*resources.TaskSingle `json:"resources,omitempty"`
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
	ProjectUUID *string `json:"project_uuid,omitempty" binding:"omitempty,uuid4" validate:"omitempty,uuid4"`
	// 備註
	Notes *string `json:"notes,omitempty"`
	// 子任務
	Subtask []*Update `json:"subtasks,omitempty"`
	// 任務分段
	Segments []*Segments `json:"segments,omitempty"`
	// 任務分段(陣列的字串型態)
	Segment *string `json:"segment,omitempty" swaggerignore:"true"`
	// 任務標示
	Indicators []*Indicators `json:"indicators,omitempty"`
	// 任務標示(陣列的字串型態)
	Indicator *string `json:"indicator,omitempty" swaggerignore:"true"`
	// 更新者
	UpdatedBy *string `json:"updated_by,omitempty" swaggerignore:"true"`
	// 資源UUID
	ResUUID *string `json:"res_uuid,omitempty" swaggerignore:"true"`
	// 角色
	Role *string `json:"role,omitempty" swaggerignore:"true"`
}

// Segments struct is used to segment the task
type Segments struct {
	// 開始日期
	StartDate time.Time `json:"start_date,omitempty"`
	// 期間
	Duration float64 `json:"duration,omitempty"`
}

// Indicators struct is used to indicator the task
type Indicators struct {
	// 日期
	Date time.Time `json:"date,omitempty"`
	// 名稱
	Name string `json:"name,omitempty"`
	// 工具提示
	ToolTip string `json:"tooltip,omitempty"`
	//
	IconClass string `json:"iconClass,omitempty"`
}

// Import struct is used to import the task file
type Import struct {
	// CSV檔案
	CSVFile *csv.Reader `swaggerignore:"true"`
	// Base64
	Base64 string `json:"base64,omitempty" binding:"required,base64" validate:"required,base64"`
	// 專案UUID
	ProjectUUID string `json:"project_uuid,omitempty" binding:"required,uuid4" validate:"required,uuid4"`
	// 檔案類型 1:gantt project 2:saas pmi
	FileType int64 `json:"file_type,omitempty" binding:"required" validate:"required"`
	// 創建者
	CreatedBy string `json:"created_by,omitempty" binding:"required,uuid4" validate:"required,uuid4" swaggerignore:"true"`
	// 資源UUID
	ResUUID *string `json:"res_uuid,omitempty" swaggerignore:"true"`
	// 角色
	Role *string `json:"role,omitempty" swaggerignore:"true"`
}

// ProjectIDs struct is used to get multiple project data
type ProjectIDs struct {
	// 多筆
	Projects []*string `json:"projects"`
	// 使用者ID
	UserID *string `json:"user_id,omitempty" swaggerignore:"true"`
	// 資源UUID
	ResUUID *string `json:"res_uuid,omitempty" swaggerignore:"true"`
	// 角色
	Role *string `json:"role,omitempty" swaggerignore:"true"`
	// 搜尋欄位
	Filter `json:"filter"`
}

// DeletedTaskUUIDs struct is used to delete multiple task data
type DeletedTaskUUIDs struct {
	// 多筆
	Tasks []*string `json:"tasks,omitempty"`
	// 專案UUID
	ProjectUUID *string `json:"project_uuid,omitempty" binding:"required,uuid4" validate:"required,uuid4"`
	// 資源UUID
	ResUUID *string `json:"res_uuid,omitempty" swaggerignore:"true"`
	// 角色
	Role *string `json:"role,omitempty" swaggerignore:"true"`
}
