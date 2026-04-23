package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"hris-backend/config/env"
	"hris-backend/config/log"
	"hris-backend/internal/redis"
	"hris-backend/internal/repository"
	"hris-backend/internal/struct/dto"
	"hris-backend/internal/utils"
)

type AuthService interface {
	Login(ctx context.Context, req dto.LoginReq) (dto.LoginRes, error)
	Refresh(ctx context.Context, refreshToken string) (dto.LoginRes, error)
	Logout(ctx context.Context, accessToken, refreshToken string) error
}

type authService struct {
	repo  repository.AuthRepository
	redis redis.Redis
}

func NewAuthService(repo repository.AuthRepository, redis redis.Redis) AuthService {
	return &authService{
		repo:  repo,
		redis: redis,
	}
}

func (s *authService) Login(ctx context.Context, req dto.LoginReq) (dto.LoginRes, error) {
	timeNow := utils.NowWIB()

	account, err := s.repo.GetAccountByEmail(ctx, nil, req.Email)
	if err != nil {
		return dto.LoginRes{}, err
	}

	if !utils.IsPasswordMatch(account.Password, req.Password) {
		return dto.LoginRes{}, errors.New("invalid password")
	}

	var result dto.LoginRes

	employee, err := s.repo.GetEmployeeByID(ctx, nil, account.ID)
	if err != nil {
		return dto.LoginRes{}, err
	}

	if len(employee.Permissions) > 0 {
		if err := json.Unmarshal(employee.Permissions, &result.Permissions); err != nil {
			return result, fmt.Errorf("failed to unmarshal permissions: %w", err)
		}
	}

	employee.Permissions = nil
	result.Account = employee

	nonce := utils.GenerateRandomString(redis.TokenLen)
	refresh := utils.GenerateRandomString(redis.TokenLen)

	tokenPayload := &dto.Token{
		Account:     employee,
		Permissions: result.Permissions,
		Email:       account.Email,
		IssuedAt:    fmt.Sprintf("%d", timeNow.Unix()),
		Expires:     fmt.Sprintf("%d", timeNow.Add(redis.TokenAuthRefreshExp).Unix()),
		Issuer:      "hris-backend",
		Subject:     fmt.Sprintf("%d", account.ID),
		Nonce:       nonce,
		Refresh:     refresh,
		Audience:    env.Cfg.Server.ClientURL,
	}

	token, err := redis.SetSession(ctx, s.redis, tokenPayload, redis.TokenAuthSessionExp)
	if err != nil {
		return result, fmt.Errorf("failed to set session: %w", err)
	}

	if err := redis.SetRefreshToken(ctx, s.redis, refresh, tokenPayload); err != nil {
		return result, fmt.Errorf("failed to set refresh session: %w", err)
	}

	result.Token = token
	result.Refresh = refresh

	if err := s.repo.UpdateAccountLastLogin(ctx, nil, timeNow, account.ID); err != nil {
		log.Error(fmt.Sprintf("failed to update last login: %v", err))
	} else {
		result.Account.LastLoginAt = &timeNow
	}

	log.Debug("login", map[string]any{
		"Exp": tokenPayload.Expires,
		"Token": token,
		"Refresh": refresh,
	})

	return result, nil
}

func (s *authService) Refresh(ctx context.Context, refreshToken string) (dto.LoginRes, error) {
	timeNow := utils.NowWIB()

	tokenData, err := redis.GetRefreshToken(ctx, s.redis, refreshToken)
	if err != nil {
		return dto.LoginRes{}, fmt.Errorf("invalid or expired refresh token: %w", err)
	}

	nonce := utils.GenerateRandomString(redis.TokenLen)
	newRefresh := utils.GenerateRandomString(redis.TokenLen)

	newTokenPayload := &dto.Token{
		Account:     tokenData.Account,
		Permissions: tokenData.Permissions,
		Email:       tokenData.Account.Email,
		IssuedAt:    fmt.Sprintf("%d", timeNow.Unix()),
		Expires:     fmt.Sprintf("%d", timeNow.Add(redis.TokenAuthRefreshExp).Unix()),
		Issuer:      "hris-backend",
		Subject:     fmt.Sprintf("%d", tokenData.Account.AccountID),
		Nonce:       nonce,
		Refresh:     newRefresh,
		Audience:    env.Cfg.Server.ClientURL,
	}

	newAccessToken, err := redis.SetSession(ctx, s.redis, newTokenPayload, redis.TokenAuthSessionExp)
	if err != nil {
		return dto.LoginRes{}, fmt.Errorf("failed to set session: %w", err)
	}

	if err := redis.SetRefreshToken(ctx, s.redis, newRefresh, newTokenPayload); err != nil {
		return dto.LoginRes{}, fmt.Errorf("failed to set refresh session: %w", err)
	}
	redis.DelRefreshToken(ctx, s.redis, refreshToken)

	return dto.LoginRes{
		Token:       newAccessToken,
		Refresh:     newRefresh,
		Account:     tokenData.Account,
		Permissions: tokenData.Permissions,
	}, nil
}

func (s *authService) Logout(ctx context.Context, accessToken, refreshToken string) error {
	redis.DelToken(ctx, s.redis, accessToken)
	redis.DelRefreshToken(ctx, s.redis, refreshToken)
	return nil
}
