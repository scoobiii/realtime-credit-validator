package auth

import (
    "errors"
    "fmt"
    "strings"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
)

type JWTConfig struct {
    SecretKey     string
    Issuer        string
    Audience      string
    TokenValidity time.Duration
}

type Claims struct {
    UserID string   `json:"user_id"`
    Scopes []string `json:"scopes"`
    jwt.RegisteredClaims
}

func NewJWTConfig(secret, issuer, audience string, validity time.Duration) *JWTConfig {
    return &JWTConfig{
        SecretKey:     secret,
        Issuer:        issuer,
        Audience:      audience,
        TokenValidity: validity,
    }
}

func (j *JWTConfig) GenerateToken(userID string, scopes []string) (string, error) {
    claims := Claims{
        UserID: userID,
        Scopes: scopes,
        RegisteredClaims: jwt.RegisteredClaims{
            Issuer:    j.Issuer,
            Audience:  []string{j.Audience},
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.TokenValidity)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Subject:   userID,
        },
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(j.SecretKey))
}

func (j *JWTConfig) ValidateToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return []byte(j.SecretKey), nil
    })
    if err != nil {
        return nil, err
    }
    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }
    return nil, errors.New("invalid token")
}

func JWTAuthMiddleware(jwtConfig *JWTConfig) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.AbortWithStatusJSON(401, gin.H{"error": "missing authorization header"})
            return
        }
        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
            c.AbortWithStatusJSON(401, gin.H{"error": "invalid authorization format"})
            return
        }
        claims, err := jwtConfig.ValidateToken(parts[1])
        if err != nil {
            c.AbortWithStatusJSON(401, gin.H{"error": "invalid or expired token"})
            return
        }
        // Verifica scope necessário
        hasScope := false
        for _, scope := range claims.Scopes {
            if scope == "credit:write" {
                hasScope = true
                break
            }
        }
        if !hasScope {
            c.AbortWithStatusJSON(403, gin.H{"error": "insufficient scope"})
            return
        }
        c.Set("user_id", claims.UserID)
        c.Next()
    }
}
