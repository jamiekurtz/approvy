package main

import (
	"github.com/jinzhu/gorm"
	"strconv"
	"time"
)

func (m *BaseModel) IDstr() string {
	return strconv.FormatUint(uint64(m.ID), 10)
}

type BaseModel struct {
	gorm.Model
}

type Request struct {
	BaseModel
	From      string
	To        string
	Message   string
	ExpiresAt time.Time
	Approved  bool
	Responses []Response
}

type Response struct {
	BaseModel
	RequestID uint
	Approved  bool
}
