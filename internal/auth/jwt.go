package auth

import (
	"strconv"
	"time"

	"HYH-Blog-Gin/internal/config"

	"github.com/golang-jwt/jwt/v5"
)

type JWTService struct {
	secret string
	expiry time.Duration
}

func NewJWTService(cfg *config.Config) *JWTService {
	return &JWTService{
		secret: cfg.JWT.Secret,
		expiry: time.Duration(cfg.JWT.Expiry) * time.Hour,
	}
}

// GenerateToken 生成 JWT Token
// userID 是用户的唯一标识符，通常是数据库中的主键 ID。
// 返回生成的 JWT Token 字符串或错误。
func (s *JWTService) GenerateToken(userID uint) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   strconv.FormatUint(uint64(userID), 10),
		Issuer:    "HYH-Blog-Gin",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.expiry)),
		NotBefore: jwt.NewNumericDate(time.Now()),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.secret))
}

// ParseToken 解析 JWT Token
// tokenStr 是要解析的 JWT Token 字符串。
// 返回用户 ID 或错误。如果 Token 无效或解析失败，将返回错误。
func (s *JWTService) ParseToken(tokenStr string) (uint, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.secret), nil
	})
	if err != nil {
		return 0, err
	}
	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		if claims.Subject == "" {
			return 0, jwt.ErrTokenInvalidClaims
		}
		uid, convErr := strconv.ParseUint(claims.Subject, 10, 64)
		if convErr != nil {
			return 0, jwt.ErrTokenInvalidClaims
		}
		return uint(uid), nil
	}
	return 0, jwt.ErrTokenInvalidClaims
}
