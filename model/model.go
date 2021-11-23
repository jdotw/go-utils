package model

import (
	"database/sql"
	"time"
)

type ID struct {
	ID string `json:"id" gorm:"primaryKey;unique;type:uuid;default:uuid_generate_v4();"`
}

type Timestamps struct {
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
	DeletedAt sql.NullTime `gorm:"index"`
}

type Defaults struct {
	ID        string       `json:"id" gorm:"primaryKey;unique;type:uuid;default:uuid_generate_v4();"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
	DeletedAt sql.NullTime `gorm:"index"`
}
