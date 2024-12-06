# Passkey <> Solana

```mermaid
graph LR
    A[Client] --> B[Relying Party]
    B --> C[Solana Program]
    A --> C
```

```mermaid
graph LR
A[Challenge] --> B[WebAuthn Sign]
B --> C[P-256 Signature]
C --> D[Bridge Contract]
D --> E[PDA Authority]
```

```mermaid
graph TD
    A[User Registration] -->| Generate WebAuthn Keys| B[Client]
    B -->| Public Key| C[Relying Party]
    C -->| Store PublicKey| D[Solana Program]
    D -->| Create PDA| E[Authority Account]
```

```mermaid
graph TD
    A[User] -->| Request Auth| B[Relying Party]
    B -->| Generate Challenge| C[Client]
    C -->| Sign with Passkey| D[Relying Party]
    D -->| Verify Signature| E[WebAuthn]
    D -->| Create Solana TX| F[Solana Program]
```

## API

### Registration

```json
POST /register/initiate
{
    "username": "string"
}

POST /register/finish/{userId}
{
    // WebAuthn attestation response
    "id": "base64",
    "rawId": "base64",
    "response": {
        "clientDataJSON": "base64",
        "attestationObject": "base64"
    },
    "type": "public-key"
}
```

### Authentication

```json
POST /login/initiate
{
    "username": "string"
}

POST /login/finish
{
    // WebAuthn assertion response
    "id": "base64",
    "rawId": "base64",
    "response": {
        "clientDataJSON": "base64",
        "authenticatorData": "base64",
        "signature": "base64",
        "userHandle": "base64"
    },
    "type": "public-key"
}
```

## Components

### Database Models

* `User`: stores user info and WebAuthn ID
* `PublicKeyCredential`: stores credential data

The passkey is generated during the finishRegistration handler when the browser responds with the attestation after the user approves the credential creation.

