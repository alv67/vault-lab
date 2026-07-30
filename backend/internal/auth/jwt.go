package auth

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/rs/zerolog/log"
)

type contextKey string

const UserContextKey contextKey = "user"

type Claims struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	TokenType string    `json:"token_type"`
}

type JWTAuth struct {
	secret      string
	accessTTL   time.Duration
	refreshTTL  time.Duration
}

func NewJWTAuth(secret string, accessTTL, refreshTTL time.Duration) *JWTAuth {
	return &JWTAuth{
		secret:     secret,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

func (a *JWTAuth) GenerateAccessToken(userID uuid.UUID, email, role string) (string, error) {
	tok, err := jwt.NewBuilder().
		Subject(userID.String()).
		Claim("email", email).
		Claim("role", role).
		Claim("token_type", "access").
		IssuedAt(time.Now()).
		Expiration(time.Now().Add(a.accessTTL)).
		JwtID(uuid.New().String()).
		Build()
	if err != nil {
		return "", err
	}
	signed, err := jwt.Sign(tok, jwt.WithKey(jwa.HS256, []byte(a.secret)))
	return string(signed), err
}

func (a *JWTAuth) GenerateRefreshToken(userID uuid.UUID) (string, error) {
	tok, err := jwt.NewBuilder().
		Subject(userID.String()).
		Claim("token_type", "refresh").
		IssuedAt(time.Now()).
		Expiration(time.Now().Add(a.refreshTTL)).
		JwtID(uuid.New().String()).
		Build()
	if err != nil {
		return "", err
	}
	signed, err := jwt.Sign(tok, jwt.WithKey(jwa.HS256, []byte(a.secret)))
	return string(signed), err
}

func (a *JWTAuth) ValidateToken(tokenString string) (*Claims, error) {
	tok, err := jwt.Parse(
		[]byte(tokenString),
		jwt.WithKey(jwa.HS256, []byte(a.secret)),
		jwt.WithValidate(true),
	)
	if err != nil {
		return nil, err
	}

	userID, err := uuid.Parse(tok.Subject())
	if err != nil {
		return nil, err
	}

	claims := &Claims{
		UserID:    userID,
		TokenType: "",
	}

	if email, ok := tok.Get("email"); ok {
		claims.Email = email.(string)
	}
	if role, ok := tok.Get("role"); ok {
		claims.Role = role.(string)
	}
	if tokenType, ok := tok.Get("token_type"); ok {
		claims.TokenType = tokenType.(string)
	}

	return claims, nil
}

func (a *JWTAuth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := extractToken(r)
		if tokenString == "" {
			http.Error(w, `{"error":"missing authorization header"}`, http.StatusUnauthorized)
			return
		}

		claims, err := a.ValidateToken(tokenString)
		if err != nil {
			log.Warn().Err(err).Msg("invalid token")
			http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
			return
		}

		if claims.TokenType != "access" {
			http.Error(w, `{"error":"invalid token type"}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func extractToken(r *http.Request) string {
	bearer := r.Header.Get("Authorization")
	if len(bearer) > 7 && bearer[:7] == "Bearer " {
		return bearer[7:]
	}
	return ""
}

func GetClaims(ctx context.Context) *Claims {
	claims, ok := ctx.Value(UserContextKey).(*Claims)
	if !ok {
		return nil
	}
	return claims
}
