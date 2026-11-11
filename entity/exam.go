package entity

import "time"

type Exam struct {
	Base
	Code             string        `gorm:"type:varchar(50);uniqueIndex;not null"`
	Name             string        `gorm:"type:varchar(100);not null"`
	Subject          string        `gorm:"type:varchar(100);not null"`
	DescriptionHTML  string        `gorm:"type:text;not null"`
	DurationMinutes  int           `gorm:"not null"`
	ShowResult       bool          `gorm:"not null;default:false"`
	Status           string        `gorm:"type:varchar(20);not null;default:'draft';index"`
	CurrentVersionID *string       `gorm:"type:char(26)"`
	Versions         []ExamVersion `gorm:"foreignKey:ExamID"`
	Sessions         []ExamSession `gorm:"foreignKey:ExamID"`
}

type ExamVersion struct {
	Base
	ExamID           string `gorm:"type:char(26);not null;index;uniqueIndex:idx_exam_version_number"`
	VersionNumber    int    `gorm:"not null;uniqueIndex:idx_exam_version_number"`
	ShuffleQuestions bool   `gorm:"not null;default:false"`
	ShuffleOptions   bool   `gorm:"not null;default:false"`
	Status           string `gorm:"type:varchar(20);not null;default:'draft';index"`
	PublishedAt      *time.Time
	Questions        []Question `gorm:"foreignKey:ExamVersionID"`
}

type ExamSession struct {
	Base
	ExamID      string           `gorm:"type:char(26);not null;index"`
	Name        string           `gorm:"type:varchar(50);not null"`
	StartsAt    time.Time        `gorm:"not null;index"`
	EndsAt      time.Time        `gorm:"not null;index"`
	Status      string           `gorm:"type:varchar(20);not null;default:'scheduled';index"`
	Assignments []ExamAssignment `gorm:"foreignKey:ExamSessionID"`
}

type Question struct {
	Base
	ExamVersionID string           `gorm:"type:char(26);not null;index"`
	Code          string           `gorm:"type:varchar(50);index"`
	ContentHTML   string           `gorm:"type:text;not null"`
	Weight        float64          `gorm:"not null;default:1"`
	DisplayOrder  int              `gorm:"not null;default:0"`
	Status        string           `gorm:"type:varchar(20);not null;default:'active';index"`
	Options       []QuestionOption `gorm:"foreignKey:QuestionID"`
}

type QuestionOption struct {
	Base
	QuestionID   string `gorm:"type:char(26);not null;index;uniqueIndex:idx_question_option_key"`
	OptionKey    string `gorm:"type:varchar(4);not null;uniqueIndex:idx_question_option_key"`
	ContentHTML  string `gorm:"type:text;not null"`
	IsCorrect    bool   `gorm:"not null;default:false"`
	DisplayOrder int    `gorm:"not null;default:0"`
}

type Candidate struct {
	Base
	ExamNumber   string           `gorm:"column:exam_number;type:varchar(100);uniqueIndex;not null"`
	PasswordHash string           `gorm:"type:varchar(255);not null" json:"-"`
	Name         string           `gorm:"type:varchar(100);not null"`
	Gender       string           `gorm:"type:varchar(1);not null"`
	School       string           `gorm:"type:varchar(100);not null"`
	ClassName    string           `gorm:"column:class_name;type:varchar(50);not null"`
	Status       string           `gorm:"type:varchar(20);not null;default:'active';index"`
	Assignments  []ExamAssignment `gorm:"foreignKey:CandidateID"`
}

type ExamAssignment struct {
	Base
	CandidateID   string    `gorm:"type:char(26);not null;index;uniqueIndex:idx_assignment_identity"`
	ExamID        string    `gorm:"type:char(26);not null;index;uniqueIndex:idx_assignment_identity"`
	ExamSessionID string    `gorm:"type:char(26);not null;index;uniqueIndex:idx_assignment_identity"`
	Status        string    `gorm:"type:varchar(20);not null;default:'assigned';index;uniqueIndex:idx_assignment_identity"`
	Attempts      []Attempt `gorm:"foreignKey:AssignmentID"`
}

type Attempt struct {
	Base
	ExamID        string            `gorm:"column:exam_id;->"`
	AssignmentID  string            `gorm:"type:char(26);not null;index"`
	AttemptNumber int               `gorm:"not null;default:1"`
	StartedAt     *time.Time        `gorm:"index"`
	DeadlineAt    *time.Time        `gorm:"index"`
	SubmittedAt   *time.Time        `gorm:"index"`
	Status        string            `gorm:"type:varchar(20);not null;default:'not_started';index"`
	CorrectCount  int               `gorm:"not null;default:0"`
	Score         float64           `gorm:"not null;default:0"`
	Questions     []AttemptQuestion `gorm:"foreignKey:AttemptID"`
	Answers       []AttemptAnswer   `gorm:"foreignKey:AttemptID"`
}

type AttemptQuestion struct {
	Base
	AttemptID           string `gorm:"type:char(26);not null;index;uniqueIndex:idx_attempt_question;uniqueIndex:idx_attempt_sequence"`
	QuestionID          string `gorm:"type:char(26);not null;index;uniqueIndex:idx_attempt_question"`
	SequenceNumber      int    `gorm:"not null;uniqueIndex:idx_attempt_sequence"`
	OptionOrderJSON     string `gorm:"type:jsonb;not null"`
	SnapshotContentJSON string `gorm:"type:jsonb;not null"`
}

type AttemptAnswer struct {
	Base
	AttemptID         string     `gorm:"type:char(26);not null;index"`
	AttemptQuestionID string     `gorm:"type:char(26);not null;uniqueIndex"`
	SelectedOptionKey string     `gorm:"type:varchar(4)"`
	AnsweredAt        *time.Time `gorm:"index"`
	IsCorrect         bool       `gorm:"not null;default:false"`
}

type AuditLog struct {
	Base
	ActorType    string `gorm:"type:varchar(30);not null;index"`
	ActorID      string `gorm:"type:char(26);index"`
	Action       string `gorm:"type:varchar(100);not null;index"`
	ResourceType string `gorm:"type:varchar(50);not null;index"`
	ResourceID   string `gorm:"type:char(26);index"`
	MetadataJSON string `gorm:"type:jsonb"`
}
