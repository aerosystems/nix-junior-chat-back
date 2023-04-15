package models

import (
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

// TokenDetails is the structure which holds data with JWT tokens
type TokenDetails struct {
	AccessToken  string
	RefreshToken string
	AccessUuid   uuid.UUID
	RefreshUuid  uuid.UUID
	AtExpires    int64
	RtExpires    int64
}

type AccessTokenClaims struct {
	AccessUUID string `json:"accessUuid"`
	UserID     int    `json:"userId"`
	Exp        int    `json:"exp"`
	jwt.StandardClaims
}

type RefreshTokenClaims struct {
	RefreshUUID string `json:"refreshUuid"`
	UserID      int    `json:"userId"`
	Exp         int    `json:"exp"`
	jwt.StandardClaims
}

type AccessTokenCache struct {
	UserID      int    `json:"userId"`
	RefreshUUID string `json:"refreshUuid"`
}

type TokensRepository interface {
	DropCacheKey(UUID string) error
	CreateCacheKey(userID int, td *TokenDetails) error
	GetCacheValue(UUID string) (*string, error)
	CreateToken(userid int) (*TokenDetails, error)
	DecodeRefreshToken(tokenString string) (*RefreshTokenClaims, error)
	DecodeAccessToken(tokenString string) (*AccessTokenClaims, error)
	DropCacheTokens(accessTokenClaims AccessTokenClaims) error
}
