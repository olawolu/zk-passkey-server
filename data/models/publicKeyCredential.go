package models

import (
	"fmt"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type PublicKeyCredential struct {
	ID                    uuid.UUID `gorm:"primaryKey"`
	UserID                uuid.UUID
	Name                  string
	PasskeyUserID         string // represented by user.id in registration options
	PublicKey             string
	AttestationType       string
	Transports            pq.StringArray
	CredentialFlags       CredentialFlags       `gorm:"foreignKey:PublicKeyCredentialId;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Authenticator         Authenticator         `gorm:"foreignKey:PublicKeyCredentialId;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	CredentialAttestation CredentialAttestation `gorm:"foreignKey:PublicKeyCredentialId;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	CreatedAt             time.Time
	UpdatedAt             time.Time
	DeletedAt             gorm.DeletedAt `gorm:"index"`
}

type CredentialFlags struct {
	gorm.Model
	PublicKeyCredentialId uuid.UUID
	UserPresent           bool `json:"userPresent"`

	// Flag UV indicates the user performed verification.
	UserVerified bool `json:"userVerified"`

	// Flag BE indicates the credential is able to be backed up and/or sync'd between devices. This should NEVER change.
	BackupEligible bool `json:"backupEligible"`

	// Flag BS indicates the credential has been backed up and/or sync'd. This value can change but it's recommended
	// that RP's keep track of this value.
	BackupState bool `json:"backupState"`
}

type Authenticator struct {
	gorm.Model
	PublicKeyCredentialId uuid.UUID

	// The AAGUID of the authenticator. An AAGUID is defined as an array containing the globally unique
	// identifier of the authenticator model being sought.
	AAGUID []byte `json:"AAGUID"`

	// SignCount -Upon a new login operation, the Relying Party compares the stored signature counter value
	// with the new signCount value returned in the assertion’s authenticator data. If this new
	// signCount value is less than or equal to the stored value, a cloned authenticator may
	// exist, or the authenticator may be malfunctioning.
	SignCount uint32 `json:"signCount"`

	// CloneWarning - This is a signal that the authenticator may be cloned, i.e. at least two copies of the
	// credential private key may exist and are being used in parallel. Relying Parties should incorporate
	// this information into their risk scoring. Whether the Relying Party updates the stored signature
	// counter value in this case, or not, or fails the authentication ceremony or not, is Relying Party-specific.
	CloneWarning bool `json:"cloneWarning"`

	// Attachment is the authenticatorAttachment value returned by the request.
	Attachment protocol.AuthenticatorAttachment `json:"attachment"`
}

type CredentialAttestation struct {
	gorm.Model
	PublicKeyCredentialId uuid.UUID
	ClientDataJSON        []byte `json:"clientDataJSON"`
	ClientDataHash        []byte `json:"clientDataHash"`
	AuthenticatorData     []byte `json:"authenticatorData"`
	PublicKeyAlgorithm    int64  `json:"publicKeyAlgorithm"`
	Object                []byte `json:"object"`
}

func CreateNewCredentials(db *gorm.DB, creds PublicKeyCredential) error {
	if err := db.Create(&creds).Error; err != nil {
		return err
	}
	return nil
}

func FetchUserCredentials(db *gorm.DB, userId uuid.UUID) ([]PublicKeyCredential, error) {
	var user User
	if err := db.Model(&User{}).Preload("PublicKeyCredentials").First(&user, userId).Error; err != nil {
		return nil, fmt.Errorf("error fetching user credentials: %v", err)
	}
	return user.PublicKeyCredentials, nil
}

func FetchUserCredential(db *gorm.DB, userId, credId uuid.UUID) (*PublicKeyCredential, error) {
	var credential PublicKeyCredential
	if err := db.First(&credential, credId).Error; err != nil {
		return nil, fmt.Errorf("error fetching credential: %v", err)
	}
	return &credential, nil
}

func UpdateCredentials(db *gorm.DB, id uuid.UUID, cred PublicKeyCredential) (*PublicKeyCredential, error) {
	if err := db.Save(&cred).Error; err != nil {
		return nil, fmt.Errorf("error updating credentials: %v", err)
	}
	return &cred, nil
}
