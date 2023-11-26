package section

import (
	"time"
)

type StartEnd struct {
	// 開始時間
	StartAt *time.Time `json:"start_at,omitempty" form:"start_at"`
	// 結束時間
	EndAt *time.Time `json:"end_at,omitempty" form:"end_at"`
}

type TimeAt struct {
	// 創建時間
	CreatedAt *time.Time `json:"created_at,omitempty"`
	// 更新時間
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
	// 刪除時間
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type ManagementExclusive struct {
	// 刪除的開始時間
	DelStartAt *time.Time `json:"del_start_at,omitempty" form:"del_start_at"`
	// 刪除的結束時間
	DelEndAt *time.Time `json:"del_end_at,omitempty" form:"del_end_at"`
}
