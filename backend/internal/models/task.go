package models

const (
	TaskStatusNew        = "new"
	TaskStatusInProgress = "in_progress"
	TaskStatusDone       = "done"
	TaskStatusCancelled  = "cancelled"
)

type Task struct {
	BaseModel
	Title       string `gorm:"size:255" json:"title"`
	Description string `gorm:"type:text" json:"description"`
	AuthorID    *uint  `gorm:"index" json:"author_id"`
	ExecutorID  *uint  `gorm:"index" json:"executor_id"`
	Photo       string `gorm:"size:512" json:"photo"`
	Status      string `gorm:"size:32;default:'new';index" json:"status"`

	Author   *User `gorm:"foreignKey:AuthorID" json:"-"`
	Executor *User `gorm:"foreignKey:ExecutorID" json:"-"`
}
