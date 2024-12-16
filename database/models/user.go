package models

import (
	"encoding/base64"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
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

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	u.ID = uuid.New()
	return
}
// PublicKeyCredential []PublicKeyCredential `gorm:"foreignKey:PasskeyUserID"`

func CreateNewUser(db *gorm.DB, user User) error {
	if err := db.Create(&user).Error; err != nil {
		return err
	}
	return nil
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

// WebAuthnID provides the user handle of the user account. A user handle is an opaque byte sequence with a maximum
// size of 64 bytes, and is not meant to be displayed to the user.
//
// To ensure secure operation, authentication and authorization decisions MUST be made on the basis of this id
// member, not the displayName nor name members. See Section 6.1 of [RFC8266].
//
// It's recommended this value is completely random and uses the entire 64 bytes.
//
// Specification: §5.4.3. User Account Parameters for Credential Generation (https://w3c.github.io/webauthn/#dom-publickeycredentialuserentity-id)
func (u User) WebAuthnID() []byte {
	uid, _ := base64.RawURLEncoding.DecodeString(u.PasskeyUserID)
	return uid
}

// WebAuthnName provides the name attribute of the user account during registration and is a human-palatable name for the user
// account, intended only for display. For example, "Alex Müller" or "田中倫". The Relying Party SHOULD let the user
// choose this, and SHOULD NOT restrict the choice more than necessary.
//
// Specification: §5.4.3. User Account Parameters for Credential Generation (https://w3c.github.io/webauthn/#dictdef-publickeycredentialuserentity)
func (u User) WebAuthnName() string {
	return u.Username
}

// WebAuthnDisplayName provides the name attribute of the user account during registration and is a human-palatable
// name for the user account, intended only for display. For example, "Alex Müller" or "田中倫". The Relying Party
// SHOULD let the user choose this, and SHOULD NOT restrict the choice more than necessary.
//
// Specification: §5.4.3. User Account Parameters for Credential Generation (https://www.w3.org/TR/webauthn/#dom-publickeycredentialuserentity-displayname)
func (u User) WebAuthnDisplayName() string {
	return u.Username
}

// WebAuthnCredentials provides the list of Credential objects owned by the user.
func (u User) WebAuthnCredentials() []webauthn.Credential {
	return nil
}

// func (u *User) AddCredential(credential *webauthn.Credential) {

// }
func (u *User) UpdateCredential(credential *webauthn.Credential) {

}
