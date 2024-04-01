package models

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type Database struct {
	DB *gorm.DB
}

func (d *Database) Tx() (*Database, func(*error)) {
	tx := d.DB.Begin().Debug()

	return &Database{
		DB: tx,
	}, func (err *error)  {
		r := recover()

		if err != nil && *err != nil {
			tx.Rollback()
			log.Debug().Msg("rollback")
		} else if err != nil && *err == nil {
			tx.Commit()
			log.Debug().Msg("commit")
		} else {
			tx.Commit()
			log.Debug().Msg("commit")
		}

		if r != nil {
			panic(r)
		}
	}
}

func (d *Database) UserHasUnacceptedMap(userID string) (bool, error) {
	m := &Map{}
	res := d.DB.
		Order("created_at DESC").
		Where(&Map{
			MapperID: userID,
		}).
		First(m)
	if res.Error != nil && errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return false, nil
	}

	if res.Error != nil {
		return false, res.Error
	}

	return m.Status == MapStatus_WaitingToAccept, nil
}

func (d *Database) UserCanUpdateMap(userID string, channelID string) (bool, error) {
	m := &Map{} 
	res := d.DB.
		Order("created_at DESC").
		Where(&Map{
			MapperID: userID,
		}).
		First(m)
	if res.Error != nil && errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return false, nil
	}

	if res.Error != nil {
		return false, res.Error
	}


	return (m.Status == MapStatus_Testing) && (m.TestingChannelID != nil) && (*m.TestingChannelID == channelID), nil 
}

func (d *Database) IsTestingChannel(channelID string) (bool, error) {
	m := &TestingChannel{}
	res := d.DB.
		Order("created_at DESC").
		Where(&TestingChannel{
			ChannelID: channelID,
		}).
		First(m)
	if res.Error != nil && errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return false, nil
	}

	if res.Error != nil {
		return false, res.Error
	}

	return true, nil
}

func (d *Database) CreateMap(name string, mapperID string, file []byte, screenshot []byte) (*Map, error) {
	m := &Map{
		Model:            Model{ID: uuid.NewString()},
		Name:             name,
		MapperID:         mapperID,
		TestingChannelID: nil,
		Status:           MapStatus_WaitingToAccept,
		File:             file,
		Screenshot:       screenshot,
	}
	return m, d.DB.Create(m).Error
}

func (d *Database) UpdateMap(id string, file []byte, screenshot []byte) error {
	return d.DB.
		Model(&Map{}).
		Where(&Map{
			Model: Model{ID: id},
		}).
		Updates(&Map{
			File:       file,
			Screenshot: screenshot,
		}).Error
}

func (d *Database) GetLastUploadedMap(mapperID string) (*Map, error) {
	m := &Map{
		MapperID: mapperID,
	}
	return m, d.DB.
		Order("created_at DESC").
		Where(&Map{
			MapperID: mapperID,
		}).
		Preload("TestingChannel").
		First(&m).Error
}

func (d *Database) GetLastUploadedMapByChannelID(channelID string) (*Map, error) {
	m := &Map{}
	return m, d.DB.
		Order("created_at DESC").
		Where(&Map{
			TestingChannelID: &channelID,
		}).
		Preload("TestingChannel").
		First(&m).Error
}

func (d *Database) AcceptMap(mapID string, mapperID string, testingChannelID string) error {
	return d.DB.Model(&Map{}).
		Where(&Map{
			Model:    Model{ID: mapID},
			MapperID: mapperID,
		}).
		Updates(&Map{
			Status:           MapStatus_Testing,
			TestingChannelID: &testingChannelID,
		}).Error
}

func (d *Database) MapExists(name string) (bool, error) {
	record := &Map{}
	res := d.DB.
		Where(&Map{
			Name: name,
		}).
		First(record)

	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return false, nil
	}

	if res.Error != nil {
		return false, fmt.Errorf("cannot get first record: %w", res.Error)
	}

	return true, nil
}

func (d *Database) CreateOrGetUser(username, id string) (*User, error) {
	user := &User{
		Model: Model{ID: id},
	}

	res := d.DB.Preload("Roles").First(user)
	if res.Error != nil && !errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("cannot get user: %w", res.Error)
	} else if res.Error != nil && errors.Is(res.Error, gorm.ErrRecordNotFound) {
		user.Username = username
		return user, d.DB.Create(user).Error
	}

	return user, nil
}

func (d *Database) CreateTestingChannel(channelID string, channelName string) (*TestingChannel, error) {
	data, err := (&TestingChannelData{
		ApprovedBy: map[string]struct{}{},
		DeclinedBy: map[string]struct{}{},
	}).ToString()
	if err != nil {
		return nil, fmt.Errorf("cannot crate testing channel data")
	}

	m := &TestingChannel{
		Model:       Model{ID: channelID},
		ChannelID:   channelID,
		ChannelName: channelName,
		Data:        data,
	}
	return m, d.DB.Create(m).Error
}

func (d *Database) GetTestingChannelData(mapID string) (*TestingChannelData, error) {
	m := &Map{}
	err := d.DB.
		Where(&Map{
			Model: Model{ID: mapID},
		}).
		Preload("TestingChannel").
		First(&m).Error
	if err != nil {
		return nil, fmt.Errorf("cannot get map: %w", err)
	}

	if m.TestingChannel == nil {
		return nil, fmt.Errorf("testing channel id nil")
	}

	data, err := TestingChannelData{}.FromString(m.TestingChannel.Data)
	if err != nil {
		return nil, fmt.Errorf("cannot parse testing channel data: %w", err)
	}

	return data, nil
}

func (d *Database) UpdateTestingChannelData(mapID string, data *TestingChannelData) error {
	m := &Map{}
	err := d.DB.
		Where(&Map{
			Model: Model{ID: mapID},
		}).
		Preload("TestingChannel").
		First(&m).Error
	if err != nil {
		return fmt.Errorf("cannot get map: %w", err)
	}

	if m.TestingChannel == nil {
		return fmt.Errorf("testing channel id nil")
	}

	dataString, err := data.ToString()
	if err != nil {
		return fmt.Errorf("cannot convert testing channel data to string: %w", err)
	}

	return d.DB.
		Model(&TestingChannel{}).
		Where(&TestingChannel{
			Model: Model{ID: m.TestingChannel.ID},
		}).
		Updates(&TestingChannel{
			Data: dataString,
		}).Error
}

func (d *Database) DeleteTestingChannel(channelID string) error {
	return d.DB.Delete(&TestingChannel{
		Model: Model{ID: channelID},
	}).Error
}
