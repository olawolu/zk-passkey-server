package database

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"github.com/olawolu/zk-pass/database/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DB struct {
	*gorm.DB
}

func NewDB(dbUrl string) *DB {
	db, err := gorm.Open(postgres.Open(dbUrl), &gorm.Config{})
	if err != nil {
		panic(fmt.Errorf("error connecting to db: %v", err))
	}

	if err = db.AutoMigrate(&models.User{}, models.PublicKeyCredential{}); err != nil {
		panic(fmt.Errorf("error migrating db: %v", err))
	}
	return &DB{db}
}

func (db *DB) RegisterNewUser(username string) (*models.User, error) {
	// create 64 byte webauthn user id
	webauthnUserID := make([]byte, 64)
	_, err := rand.Read(webauthnUserID)
	if err != nil {
		return nil, fmt.Errorf("error generating webauthn user id: %v", err)
	}
	// conver webauthnUserID into a string
	webauthnUserIDStr := base64.RawURLEncoding.EncodeToString(webauthnUserID)
	newUser := models.User{
		ID:            uuid.New(),
		Username:      username,
		PasskeyUserID: webauthnUserIDStr,
	}
	err = models.CreateNewUser(db.DB, newUser)
	if err != nil {
		return nil, err
	}

	return &newUser, nil
}

func (db *DB) GetUser(id string) (*models.User, error) {
	// fetch user
	uid, err := base64.RawURLEncoding.DecodeString(id)
	if err != nil {
		return nil, fmt.Errorf("error decoding userId string: %v", err)
	}
	userId, err := uuid.ParseBytes(uid)
	if err != nil {
		return nil, fmt.Errorf("error parsing decoded userId string: %v", err)
	}
	u, err := models.FetchUser(db.DB, userId)
	if err != nil {
		return nil, fmt.Errorf("error fetching user %v from db: %v", id, err)
	}

	return u, nil
}

func (db *DB) SaveUser(user models.User) error {
	_, err := models.UpdateUser(db.DB, user)
	if err != nil {
		return fmt.Errorf("error updating user %v in db: %v", user.ID, err)
	}
	return nil
}

func (db *DB) AddCredential(credential *webauthn.Credential, userId uuid.UUID, name string) error {
	var transports []string
	credentialUid := uuid.New()
	credentialId := base64.RawURLEncoding.EncodeToString(credential.ID)
	publicKey := base64.RawURLEncoding.EncodeToString(credential.PublicKey)

	for _, v := range credential.Transport {
		transports = append(transports, string(v))
	}

	newCredential := models.PublicKeyCredential{
		ID:              credentialUid,
		Name:            name,
		UserID:          userId,
		PasskeyUserID:   credentialId,
		PublicKey:       publicKey,
		AttestationType: credential.AttestationType,
		Transports:      transports,
		CredentialFlags: models.CredentialFlags{
			PublicKeyCredentialId: credentialUid,
			UserPresent:           credential.Flags.UserPresent,
			UserVerified:          credential.Flags.UserVerified,
			BackupEligible:        credential.Flags.BackupEligible,
			BackupState:           credential.Flags.BackupState,
		},
		Authenticator: models.Authenticator{
			PublicKeyCredentialId: credentialUid,
			AAGUID:                credential.Authenticator.AAGUID,
			SignCount:             credential.Authenticator.SignCount,
			CloneWarning:          credential.Authenticator.CloneWarning,
			Attachment:            credential.Authenticator.Attachment,
		},
		CredentialAttestation: models.CredentialAttestation{
			PublicKeyCredentialId: credentialUid,
			ClientDataJSON:        credential.Attestation.ClientDataJSON,
			ClientDataHash:        credential.Attestation.ClientDataHash,
			AuthenticatorData:     credential.Attestation.AuthenticatorData,
			PublicKeyAlgorithm:    credential.Attestation.PublicKeyAlgorithm,
			Object:                credential.Attestation.Object,
		},
	}

	if err := models.CreateNewCredentials(db.DB, newCredential); err != nil {
		return fmt.Errorf("error saving credentials: %v", err)
	}
	return nil
}
