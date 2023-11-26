package work_days

import (
	"hta/internal/interactor/models/page"
	"hta/internal/interactor/models/section"
)

// Create struct is used end_date create achieves
type Create struct {
	// 工作日
	WorkWeeks []string `json:"workWeek,omitempty"`
	// 工作日(陣列的字串型態)
	WorkWeek string `json:"work_week,omitempty" swaggerignore:"true"`
	// 工作時間
	WorkingTimes []*WorkingTimes `json:"workingTime,omitempty"`
	// 工作時間(陣列的字串型態)
	WorkingTime string `json:"working_time,omitempty" swaggerignore:"true"`
	// 創建者
	CreatedBy string `json:"created_by,omitempty" binding:"required,uuid4" validate:"required,uuid4" swaggerignore:"true"`
}

// Field is structure file for search
type Field struct {
	// 表ID
	ID string `json:"id,omitempty" binding:"omitempty,uuid4" validate:"omitempty,uuid4" swaggerignore:"true"`
}

// Fields is the searched structure file (including pagination)
type Fields struct {
	// 搜尋結構檔
	Field
	// 分頁搜尋結構檔
	page.Pagination
}

// List is multiple return structure files
type List struct {
	// 多筆
	WorkDays []*struct {
		// 表ID
		ID string `json:"id,omitempty"`
		// 工作日
		WorkWeeks []string `json:"workWeek,omitempty"`
		// 工作時間
		WorkingTimes []WorkingTimes `json:"workingTime,omitempty"`
		// 創建者
		CreatedBy string `json:"created_by,omitempty"`
		// 更新者
		UpdatedBy string `json:"updated_by,omitempty"`
		// 時間戳記
		section.TimeAt
	} `json:"work_days"`
	// 分頁返回結構檔
	page.Total
}

// Single return structure file
type Single struct {
	// 表ID
	ID string `json:"id,omitempty"`
	// 工作日
	WorkWeeks []string `json:"workWeek,omitempty"`
	// 工作時間
	WorkingTimes []WorkingTimes `json:"workingTime,omitempty"`
	// 創建者
	CreatedBy string `json:"created_by,omitempty"`
	// 更新者
	UpdatedBy string `json:"updated_by,omitempty"`
	// 時間戳記
	section.TimeAt
}

// Update struct is used end_date update achieves
type Update struct {
	// 表ID
	ID string `json:"id,omitempty"  binding:"omitempty,uuid4" validate:"omitempty,uuid4" swaggerignore:"true"`
	// 工作日
	WorkWeeks []*string `json:"workWeek,omitempty"`
	// 工作日(陣列的字串型態)
	WorkWeek *string `json:"work_week,omitempty" swaggerignore:"true"`
	// 工作時間
	WorkingTimes []*WorkingTimes `json:"workingTime,omitempty"`
	// 工作時間(陣列的字串型態)
	WorkingTime *string `json:"working_time,omitempty" swaggerignore:"true"`
	// 更新者
	UpdatedBy *string `json:"updated_by,omitempty" binding:"required,uuid4" validate:"required,uuid4" swaggerignore:"true"`
}

// WorkingTimes struct is used to set working time
type WorkingTimes struct {
	// 開始時間
	StartTime float32 `json:"from,omitempty"`
	// 結束時間
	EndTime float32 `json:"to,omitempty"`
}
