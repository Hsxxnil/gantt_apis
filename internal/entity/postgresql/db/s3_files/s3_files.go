package s3_files

import (
	"hta/internal/entity/postgresql/db/users"
	"hta/internal/interactor/models/special"
	"time"

	"gorm.io/gorm"
)

// Table struct is s3_files database table struct
type Table struct {
	// 表ID
	ID string `gorm:"<-:create;column:id;type:uuid;not null;primaryKey;" json:"id"`
	// 檔案連結
	FileUrl string `gorm:"<-:create;column:file_url;type:text;" json:"file_url"`
	// 檔案名稱
	FileName string `gorm:"<-:create;column:file_name;type:text;" json:"file_name"`
	// 副檔名
	FileExtension string `gorm:"<-:create;column:file_extension;type:text;" json:"file_extension"`
	// 來源UUID
	SourceUUID string `gorm:"<-:create;column:source_uuid;type:uuid;" json:"source_uuid"`
	// 創建時間
	CreatedAt time.Time `gorm:"<-:create;column:created_at;type:TIMESTAMP;not null;" json:"created_at"`
	// 創建人
	CreatedBy string `gorm:"<-:create;column:created_by;type:uuid;" json:"created_by"`
	// create_users data
	CreatedByUsers users.Table `gorm:"foreignKey:ID;references:CreatedBy" json:"created_by_users,omitempty"`
	// 刪除時間
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;type:TIMESTAMP;" json:"deleted_at,omitempty"`
}

// Base struct is corresponding to s3_files table structure file
type Base struct {
	// 表ID
	ID *string `json:"id,omitempty"`
	// 檔案連結
	FileUrl *string `json:"file_url,omitempty"`
	// 檔案名稱
	FileName *string `json:"file_name,omitempty"`
	// 副檔名
	FileExtension *string `json:"file_extension,omitempty"`
	// 來源UUID
	SourceUUID *string `json:"source_uuid,omitempty"`
	// create_users data
	CreatedByUsers users.Base `json:"created_by_users,omitempty"`
	// 引入後端專用
	special.Base
}

func (a *Table) TableName() string {
	return "s3_files"
}
