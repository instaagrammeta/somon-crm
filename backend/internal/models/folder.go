package models

type Folder struct {
	BaseModel
	Name       string `gorm:"size:255" json:"name"`
	ParentID   uint   `gorm:"default:0;index" json:"parent_id"`
	AuthorID   *uint  `gorm:"index" json:"author_id"`
	AuthorName string `gorm:"size:255" json:"author_name"`
}

type FolderFile struct {
	BaseModel
	FolderID     uint   `gorm:"index" json:"folder_id"`
	Filename     string `gorm:"size:255" json:"filename"`
	OriginalName string `gorm:"size:255" json:"original_name"`
	FilePath     string `gorm:"size:512" json:"filepath"`
	FileType     string `gorm:"size:64" json:"filetype"`
	FileSize     int64  `json:"filesize"`
	AuthorID     *uint  `gorm:"index" json:"author_id"`
	AuthorName   string `gorm:"size:255" json:"author_name"`
}

func (FolderFile) TableName() string { return "folder_files" }
