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
	// ExecutorID keeps the *primary* executor for backward compatibility with
	// older clients and existing queries. ExecutorIDs holds the full set.
	ExecutorID  *uint     `gorm:"index" json:"executor_id"`
	ExecutorIDs UintSlice `gorm:"type:jsonb;default:'[]'" json:"executor_ids"`
	// Photo keeps the *first* photo for backward compatibility. Photos holds all
	// of them (ordered) so the UI can show a gallery and reorder/delete them.
	Photo  string      `gorm:"size:512" json:"photo"`
	Photos StringSlice `gorm:"type:jsonb;default:'[]'" json:"photos"`
	Status string      `gorm:"size:32;default:'new';index" json:"status"`

	Author   *User `gorm:"foreignKey:AuthorID" json:"-"`
	Executor *User `gorm:"foreignKey:ExecutorID" json:"-"`
}
