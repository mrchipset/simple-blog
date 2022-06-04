package blog

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Post struct {
	UUID         uuid.UUID
	Title        string `json:"title"`
	Summary      string `json:"summary"`
	Link         string `json:"link"`
	PublishState bool   `json:"public_state"`
	gorm.Model   `gorm:"embedded"`
}

type PostContent struct {
	UUID       uuid.UUID
	Content    string `json:"content"`
	CheckSum   string `json:"checksum"`
	gorm.Model `gorm:"embedded"`
}

type PostQuery struct {
	Post        Post        `json:"post"`
	PostContent PostContent `json:"content"`
}

type Pagenation struct {
	Prev  int
	Next  int
	Pages []int
}
