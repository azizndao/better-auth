package core

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Model struct {
	ID        uuid.UUID      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `                  json:"createdAt"`
	UpdatedAt time.Time      `                  json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index"      json:"-"`
}

func NewModelFromID(id uuid.UUID) (*Model, error) {
	return &Model{ID: id}, nil
}
