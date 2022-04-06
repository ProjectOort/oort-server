package account

import (
	"encoding/base64"
	"errors"
	"time"

	"github.com/ProjectOort/oort-server/conf"
	"github.com/golang-jwt/jwt/v4"
)

// compile-time interface implementation check.
var (
	_ TokenMaker     = (*tokenHelper)(nil)
	_ TokenValidator = (*tokenHelper)(nil)
)

type tokenHelper struct {
	key      []byte
	duration time.Duration
}

func newTokenHelper(cfg *conf.Account) *tokenHelper {
	if cfg.TokenKey == "" {
		panic("token key shouldn't be empty")
	}
	key, err := base64.StdEncoding.DecodeString(cfg.TokenKey)
	if err != nil {
		panic("invalid token key")
	}
	return &tokenHelper{
		key:      key,
		duration: time.Duration(cfg.TokenExpireSec) * time.Second,
	}
}

func (x *tokenHelper) MakeToken(payload map[string]interface{}) (string, error) {
	clm := make(jwt.MapClaims)
	for k, v := range payload {
		clm[k] = v
	}
	clm["exp"] = time.Now().Add(x.duration).UnixMilli()
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, clm)
	return token.SignedString(x.key)
}

func (x *tokenHelper) ValidateToken(token string) (map[string]interface{}, error) {
	t, err := jwt.ParseWithClaims(token, &jwt.MapClaims{}, func(t *jwt.Token) (interface{}, error) {
		return x.key, nil
	})
	if err != nil {
		return nil, err
	}
	if !t.Valid {
		return nil, errors.New("invaild token")
	}
	return *(t.Claims.(*jwt.MapClaims)), nil
}
