package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username string
	Roles    []Role
}

type RoleName string

var (
	RoleName_MapAcceptor RoleName = "map_acceptor"
	RoleName_Tester      RoleName = "tester"
)

type Role struct {
	gorm.Model
	UserID int
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
	gorm.Model
	Name             string
	FixedName        string
	MapperID         int
	Mapper           User
	TestingChannelID int
	TestingChannel   TestingChannel
	Status           MapStatus
	File             []byte
}

type TestingChannel struct {
	gorm.Model
	Name              string
	ApprovedByTesters int
	DeclinedByTesters int
}
