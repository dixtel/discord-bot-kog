package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type Model struct {
	ID        string `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type User struct {
	Model
	Username string
}

type MapStatus string

var (
	MapStatus_WaitingToAccept MapStatus = "waiting_to_accept"
	MapStatus_Accepted        MapStatus = "accepted"
	MapStatus_Rejected        MapStatus = "rejected"
	MapStatus_Approved        MapStatus = "approved"
	MapStatus_Declined        MapStatus = "declined"
)

type Map struct {
	Model
	FileName         string
	MapperID         string
	Mapper           User
	TestingChannelID *string         // NULL if status is 'waiting_to_accept'
	TestingChannel   *TestingChannel // NULL if status is 'waiting_to_accept'
	Status           MapStatus
	File             []byte
	Screenshot       []byte
}

type TestingChannelData struct {
	ApprovedBy map[string]struct{} // Tester ID as key
	DeclinedBy map[string]struct{} // Tester ID as key
}

type TestingChannel struct {
	Model
	ChannelID   string
	ChannelName string
	Data        string // TestingChannelData
}

func (d *TestingChannelData) ToString() (string, error) {
	b, err := json.Marshal(&d)
	return string(b), err
}

func (TestingChannelData) FromString(src string) (*TestingChannelData, error) {
	var data TestingChannelData

	err := json.Unmarshal([]byte(src), &data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

type BannedUserFromSubmission struct {
	Model
	BannedUserID string `gorm:"unique"`
	Reason       string
	ByUserID     string
}
