package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"hris-backend/config/log"
	"hris-backend/internal/struct/dto"
	"hris-backend/internal/utils"
)

const (
	TokenLen            = 12
	TokenAuthSessionExp = time.Hour * 1     // 1 jam
	TokenAuthRefreshExp = time.Hour * 24 * 30 // 30 hari

	ServerAuthMsgInvalidToken = "invalid verification token"
	ServerAuthMsgTokenExpired = "token expired, please refresh your token"
)

var (
	ServerErrInvalidToken = errors.New(ServerAuthMsgInvalidToken)
	ServerErrTokenExpired = errors.New(ServerAuthMsgTokenExpired)
)

func NewToken(email, uid, nonce string, duration time.Duration) dto.Token {
	iat := fmt.Sprintf("%d", time.Now().Unix())
	exp := fmt.Sprintf("%d", time.Now().Add(duration).Unix())

	return dto.Token{
		Subject:  uid,
		Email:    email,
		IssuedAt: iat,
		Expires:  exp,
		Nonce:    nonce,
	}
}

func PrepToken(email string, uid string) dto.Token {
	code := utils.GenerateRandomString(TokenLen)
	return NewToken(email, uid, code, TokenAuthSessionExp)
}

func SetToken(ctx context.Context, rdb Redis, t dto.Token) (string, error) {
	key := utils.GenerateRandomString(TokenLen)

	err := settoken(ctx, rdb, key, t, TokenAuthSessionExp)
	if err != nil {
		return "", fmt.Errorf("auth: error setting token %w", err)
	}

	return key, nil
}

func SetSession(ctx context.Context, rdb Redis, t *dto.Token, exp time.Duration) (string, error) {
	key := utils.GenerateRandomString(TokenLen)

	err := settoken(ctx, rdb, key, *t, exp)
	if err != nil {
		return "", fmt.Errorf("auth: error setting session %w", err)
	}

	return key, nil
}

func SetSessionCustom[T any](ctx context.Context, rdb Redis, data T, key *string, duration time.Duration) (*string, error) {
	if key == nil {
		_temp := utils.GenerateRandomString(TokenLen)
		key = &_temp
	}

	jt, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("auth: error setting token %w", err)
	}

	err = rdb.Set(ctx, *key, string(jt), duration)
	if err != nil {
		return nil, fmt.Errorf("auth: error setting session %w", err)
	}

	return key, nil
}

func settoken(ctx context.Context, rdb Redis, key string, t dto.Token, exp time.Duration) error {
	jt, err := json.Marshal(t)
	if err != nil {
		return fmt.Errorf("auth: error setting token %w", err)
	}

	err = rdb.Set(ctx, key, string(jt), exp)
	if err != nil {
		return fmt.Errorf("auth: error setting token %w", err)
	}

	return nil
}

func RetrieveToken(ctx context.Context, rdb Redis, key string) (*dto.Token, error) {
	found, err := rdb.Exists(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("auth: error retrieving token %w", err)
	}

	if found <= 0 {
		return nil, fmt.Errorf("auth: token not found")
	}

	jt, err := rdb.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("auth: error retrieving token %w", err)
	}

	var token dto.Token
	err = json.Unmarshal([]byte(jt), &token)
	if err != nil {
		return nil, fmt.Errorf("auth: error retrieving token %w", err)
	}

	return &token, nil
}

func GetData[T any](ctx context.Context, rdb Redis, key string) (*T, error) {
	found, err := rdb.Exists(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("auth: error retrieving data %w", err)
	}

	if found <= 0 {
		return nil, fmt.Errorf("auth: data not found")
	}

	jt, err := rdb.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("auth: error retrieving data %w", err)
	}

	var data T
	err = json.Unmarshal([]byte(jt), &data)
	if err != nil {
		return nil, fmt.Errorf("auth: error retrieving data %w", err)
	}

	return &data, nil
}

func GetToken(ctx context.Context, rdb Redis, key string) (*dto.Token, error) {
	jt, err := rdb.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("auth: error retrieving token %w", err)
	}

	var token dto.Token
	err = json.Unmarshal([]byte(jt), &token)
	if err != nil {
		return nil, fmt.Errorf("auth: error retrieving token %w", err)
	}

	err = CheckToken(&token)
	if err != nil {
		return nil, fmt.Errorf("auth: error checking token %w", err)
	}

	return &token, nil
}

func CheckToken(t *dto.Token) error {
	now := time.Now().Unix()
	exp, err := strconv.Atoi(t.Expires)
	if err != nil {
		return fmt.Errorf("auth: error parsing token expiricy %w", err)
	}

	if now > int64(exp) {
		return fmt.Errorf("auth: token expired")
	}

	return nil
}

func DelToken(ctx context.Context, rdb Redis, key string) error {
	_, err := rdb.Del(ctx, key)
	if err != nil {
		return fmt.Errorf("auth: error deleting token %w", err)
	}

	return nil
}

func SetRefreshToken(ctx context.Context, rdb Redis, refreshToken string, t *dto.Token) error {
	jt, err := json.Marshal(t)
	if err != nil {
		return err
	}
	return rdb.Set(ctx, "refresh:"+refreshToken, string(jt), TokenAuthRefreshExp)
}

func GetRefreshToken(ctx context.Context, rdb Redis, refreshToken string) (*dto.Token, error) {
	jt, err := rdb.Get(ctx, "refresh:"+refreshToken)
	log.Debug("token", map[string]any{
		"refresh token": "refresh:" + refreshToken,
	})
	if err != nil {
		return nil, fmt.Errorf("refresh token not found or expired")
	}
	var token dto.Token
	if err := json.Unmarshal([]byte(jt), &token); err != nil {
		return nil, err
	}
	log.Debug("token: %v", map[string]any{
		"jt":    jt,
		"token": token,
	})
	err = CheckToken(&token)
	if err != nil {
		return nil, fmt.Errorf("auth: error checking refresh token: %w", err)
	}

	return &token, nil
}

func DelRefreshToken(ctx context.Context, rdb Redis, refreshToken string) error {
	_, err := rdb.Del(ctx, "refresh:"+refreshToken)
	if err != nil {
		return fmt.Errorf("auth: error deleting refresh token %w", err)
	}
	return nil
}
