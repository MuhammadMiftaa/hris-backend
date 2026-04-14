package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"hris-backend/config/env"
	"hris-backend/internal/redis"
	"hris-backend/internal/repository"
	"hris-backend/internal/struct/dto"
	"hris-backend/internal/utils"
)

type AuthService interface {
	Login(ctx context.Context, req dto.LoginReq) (dto.LoginRes, error)
	Refresh(ctx context.Context, req dto.LoginRes) (dto.LoginRes, error)
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

	token, err := redis.SetSession(ctx, s.redis, &dto.Token{
		Account:     employee,
		Permissions: result.Permissions,
		Email:       account.Email,
		IssuedAt:    fmt.Sprintf("%d", time.Now().Unix()),
		Expires:     fmt.Sprintf("%d", time.Now().Add(redis.TokenAuthSessionExp).Unix()),
		Issuer:      "hris-backend",
		Subject:     fmt.Sprintf("%d", account.ID),
		Nonce:       nonce,
		Refresh:     refresh,
		Audience:    env.Cfg.Server.ClientURL,
	}, time.Hour*24)
	if err != nil {
		return result, fmt.Errorf("failed to set session: %w", err)
	}

	result.Token = token
	result.Refresh = refresh

	return result, nil
}

func (s *authService) Refresh(ctx context.Context, user dto.LoginRes) (dto.LoginRes, error) {
	nonce := utils.GenerateRandomString(redis.TokenLen)
	refresh := utils.GenerateRandomString(redis.TokenLen)

	token, err := redis.SetSession(ctx, s.redis, &dto.Token{
		Account:     user.Account,
		Permissions: user.Permissions,
		Email:       user.Account.Email,
		IssuedAt:    fmt.Sprintf("%d", time.Now().Unix()),
		Expires:     fmt.Sprintf("%d", time.Now().Add(redis.TokenAuthSessionExp).Unix()),
		Issuer:      "hris-backend",
		Subject:     fmt.Sprintf("%d", user.Account.AccountID),
		Nonce:       nonce,
		Refresh:     refresh,
		Audience:    env.Cfg.Server.ClientURL,
	}, time.Hour*24)
	if err != nil {
		return dto.LoginRes{}, fmt.Errorf("failed to set session: %w", err)
	}

	s.redis.Del(ctx, user.Token)
	user.Token = token
	user.Refresh = refresh

	return user, nil
}
