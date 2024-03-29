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
	Roles    []Role
}

func (u *User) HasRole(wantedRole RoleName) bool {
	for _, role := range u.Roles {
		if role.Role == wantedRole {
			return true
		}
	}
	return false
}

type RoleName string

var (
	RoleName_MapAcceptor RoleName = "map_acceptor"
	RoleName_Tester      RoleName = "tester"
)

type Role struct {
	Model
	UserID string
	User   User
	Role   RoleName
}

type MapStatus string

var (
	MapStatus_WaitingToAccept MapStatus = "waiting_to_accept"
	MapStatus_Testing         MapStatus = "testing"
	MapStatus_Accepted        MapStatus = "accepted"
	MapStatus_Declined        MapStatus = "declined"
)

type Map struct {
	Model
	Name     string
	MapperID string
	Mapper   User
	// NULL if status is 'waiting_to_accept'
	TestingChannelID *string
	// NULL if status is 'waiting_to_accept'
	TestingChannel *TestingChannel
	Status         MapStatus
	File           []byte
}

type TestingChannelData struct {
	// Tester ID as key
	ApprovedBy map[string]struct{}
	// Tester ID as key
	DeclinedBy map[string]struct{}
}

type TestingChannel struct {
	Model
	ChannelID   string
	ChannelName string
	// TestingChannelData
	Data string
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
