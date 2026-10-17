package auth

import (
	"errors"
	"fmt"
	"time"
	"uuid"

	"github.com/golang-jwt/jwt/v5"

	appmiddleware "hnu-tickets/internal/platform/middleware"
)

type Claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

type TokenManager struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

func NewTokenManager(secret string, ttl time.Duration) *TokenManager {
	return &TokenManager{secret: []byte(secret), ttl: ttl, now: time.Now}
}

func (m *TokenManager) Issue(userID uuid.UUID, role string) (string, error) {
	now := m.now().UTC()
	claims := Claims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
	if err != nil {
		return "", fmt.Errorf("sign access token: %w", err)
	}
	return token, nil
}

func (m *TokenManager) Parse(raw string) (appmiddleware.Principal, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return m.secret, nil
	}, jwt.WithExpirationRequired(), jwt.WithTimeFunc(m.now))
	if err != nil || !token.Valid {
		return appmiddleware.Principal{}, errors.New("invalid access token")
	}
	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return appmiddleware.Principal{}, errors.New("invalid token subject")
	}
	tokenID, err := uuid.Parse(claims.ID)
	if err != nil || tokenID == uuid.Nil() || userID == uuid.Nil() {
		return appmiddleware.Principal{}, errors.New("invalid token identity")
	}
	return appmiddleware.Principal{UserID: userID, Role: claims.Role, TokenID: tokenID, ExpiresAt: claims.ExpiresAt.Time}, nil
}
