package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID                   uuid.UUID `gorm:"primaryKey" `
	PasskeyUserID        string
	Username             string
	CreatedAt            time.Time             `gorm:"index;type:timestamptz;not null;default:NOW()"`
	UpdatedAt            time.Time             `gorm:"index;type:timestamptz"`
	DeletedAt            gorm.DeletedAt        `gorm:"index"`
	PublicKeyCredentials []PublicKeyCredential `gorm:"foreignKey:PasskeyUserID"`
}

// PublicKeyCredential []PublicKeyCredential `gorm:"foreignKey:PasskeyUserID"`

func CreateNewUser(db *gorm.DB, user User) (*User, error) {
	user.ID = uuid.New()
	if err := db.Create(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func FetchUser(db *gorm.DB, id uuid.UUID) (*User, error) {
	var user User
	if err := db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func UpdateUser(db *gorm.DB, user User) (*User, error) {
	if err := db.Save(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}	