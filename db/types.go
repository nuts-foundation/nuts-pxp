package db

import "gorm.io/gorm/schema"

var _ schema.Tabler = (*SQLData)(nil)

type SQLData struct {
	Id        string `gorm:"primaryKey"`
	Client    string
	Scope     string
	Verifier  string
	AuthInput string `gorm:"column:auth_input"`
}

func (s SQLData) TableName() string {
	return "data"
}
