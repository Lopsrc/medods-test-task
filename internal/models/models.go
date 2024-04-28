package models

import "time"

type Tokens struct{
	AccessToken string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type GetTokens struct{
	GUID string `json:"guid"`
}

type UpdateTokens struct{
	RefreshToken string `json:"refresh_token"`
}

type User struct {
	GUID string `bson:"guid"`
	RefreshTokenHash []byte `bson:"refresh_token_hash"`
	ExpiresAt time.Time `bson:"expiresAt"`
}
