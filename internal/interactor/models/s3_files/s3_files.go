package s3_files

import "hta/internal/interactor/models/page"

// Create struct is used to create achieves
type Create struct {
	// 檔案連結
	FileUrl string `json:"file_url,omitempty" swaggerignore:"true"`
	// 檔案名稱
	FileName string `json:"file_name,omitempty" binding:"required" validate:"required"`
	// 副檔名
	FileExtension string `json:"file_extension,omitempty" swaggerignore:"true"`
	// 來源UUID
	SourceUUID string `json:"source_uuid,omitempty" binding:"required,uuid4" validate:"required,uuid4"`
	// Base64
	Base64 string `json:"base64,omitempty" binding:"required,base64" validate:"required,base64"`
	// 創建者
	CreatedBy string `json:"created_by,omitempty" binding:"omitempty,uuid4" validate:"omitempty,uuid4" swaggerignore:"true"`
}

// Field is structure file for search
type Field struct {
	// 表ID
	ID string `json:"id,omitempty" binding:"omitempty,uuid4" validate:"omitempty,uuid4" swaggerignore:"true"`
	// 檔案連結
	FileUrl *string `json:"file_url,omitempty" form:"file_url"`
	// 檔案名稱
	FileName *string `json:"file_name,omitempty" form:"file_name"`
	// 副檔名
	FileExtension *string `json:"file_extension,omitempty" form:"file_extension"`
	// 來源UUID
	SourceUUID *string `json:"source_uuid,omitempty" form:"source_uuid"`
}

// Single is single return structure file
type Single struct {
	// 表ID
	ID string `json:"id,omitempty"`
	// 檔案連結
	FileUrl string `json:"file_url,omitempty"`
	// 檔案名稱
	FileName string `json:"file_name,omitempty"`
	// 創建者
	CreatedBy string `json:"created_by,omitempty"`
	// 創建時間
	CreatedAt string `json:"created_at,omitempty"`
}

// Fields is the searched structure file (including pagination)
type Fields struct {
	// 搜尋結構檔
	Field
	// 分頁搜尋結構檔
	page.Pagination
}
