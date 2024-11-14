package data

import "github.com/go-webauthn/webauthn/webauthn"


type User struct {
	UserId        string
	PasskeyUserId string
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
	return nil
}

// WebAuthnName provides the name attribute of the user account during registration and is a human-palatable name for the user
// account, intended only for display. For example, "Alex Müller" or "田中倫". The Relying Party SHOULD let the user
// choose this, and SHOULD NOT restrict the choice more than necessary.
//
// Specification: §5.4.3. User Account Parameters for Credential Generation (https://w3c.github.io/webauthn/#dictdef-publickeycredentialuserentity)
func (u User) WebAuthnName() string {
	return ""
}

// WebAuthnDisplayName provides the name attribute of the user account during registration and is a human-palatable
// name for the user account, intended only for display. For example, "Alex Müller" or "田中倫". The Relying Party
// SHOULD let the user choose this, and SHOULD NOT restrict the choice more than necessary.
//
// Specification: §5.4.3. User Account Parameters for Credential Generation (https://www.w3.org/TR/webauthn/#dom-publickeycredentialuserentity-displayname)
func (u User) WebAuthnDisplayName() string {
	return ""
}

// WebAuthnCredentials provides the list of Credential objects owned by the user.
func (u User) WebAuthnCredentials() []webauthn.Credential {
	return nil
}

func (u *User) AddCredential(credential *webauthn.Credential) {

}
func (u *User) UpdateCredential(credential *webauthn.Credential)  {
	
}