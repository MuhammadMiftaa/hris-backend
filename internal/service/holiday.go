package service

import (
	"context"
	"fmt"
	"time"

	"hris-backend/internal/repository"
	"hris-backend/internal/struct/dto"
	"hris-backend/internal/struct/model"
	"hris-backend/internal/utils/data"
)

type HolidayService interface {
	GetMetadata(ctx context.Context) (dto.HolidayMetadata, error)
	GetAllHolidays(ctx context.Context, params *dto.HolidayListParams) ([]dto.HolidayResponse, error)
	GetHolidayByID(ctx context.Context, id uint) (dto.HolidayResponse, error)
	CreateHoliday(ctx context.Context, req dto.CreateHolidayRequest) (dto.HolidayResponse, error)
	UpdateHoliday(ctx context.Context, id uint, req dto.UpdateHolidayRequest) (dto.HolidayResponse, error)
	DeleteHoliday(ctx context.Context, id uint) error
}

type holidayService struct {
	repo repository.HolidayRepository
}

func NewHolidayService(repo repository.HolidayRepository) HolidayService {
	return &holidayService{repo: repo}
}

var validHolidayTypes = map[string]bool{
	"national":    true,
	"joint":       true,
	"observance":  true,
	"company":     true,
}

func (s *holidayService) GetMetadata(ctx context.Context) (dto.HolidayMetadata, error) {
	branchMeta, err := s.repo.GetBranchMetadata(ctx, nil)
	if err != nil {
		return dto.HolidayMetadata{}, fmt.Errorf("get branch metadata: %w", err)
	}

	return dto.HolidayMetadata{
		HolidayTypeMeta: data.HolidayTypeMeta,
		BranchMeta:      branchMeta,
	}, nil
}

func (s *holidayService) GetAllHolidays(ctx context.Context, params *dto.HolidayListParams) ([]dto.HolidayResponse, error) {
	holidays, err := s.repo.GetAllHolidays(ctx, nil, params)
	if err != nil {
		return nil, fmt.Errorf("get all holidays: %w", err)
	}
	return holidays, nil
}

func (s *holidayService) GetHolidayByID(ctx context.Context, id uint) (dto.HolidayResponse, error) {
	holiday, err := s.repo.GetHolidayByID(ctx, nil, id)
	if err != nil {
		return dto.HolidayResponse{}, fmt.Errorf("get holiday by ID: %w", err)
	}
	return holiday, nil
}

func (s *holidayService) CreateHoliday(ctx context.Context, req dto.CreateHolidayRequest) (dto.HolidayResponse, error) {
	if req.Name == "" {
		return dto.HolidayResponse{}, fmt.Errorf("name is required")
	}
	if req.Date == "" {
		return dto.HolidayResponse{}, fmt.Errorf("date is required")
	}
	if req.Type == "" {
		return dto.HolidayResponse{}, fmt.Errorf("type is required")
	}
	if !validHolidayTypes[req.Type] {
		return dto.HolidayResponse{}, fmt.Errorf("invalid holiday type: %q", req.Type)
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return dto.HolidayResponse{}, fmt.Errorf("invalid date format (expected YYYY-MM-DD): %w", err)
	}

	holidayModel := model.Holiday{
		Name:        req.Name,
		Year:        date.Year(),
		Date:        date,
		Type:        model.HolidayTypeEnum(req.Type),
		BranchID:    req.BranchID,
		Description: req.Description,
	}

	created, err := s.repo.CreateHoliday(ctx, nil, holidayModel)
	if err != nil {
		return dto.HolidayResponse{}, fmt.Errorf("create holiday: %w", err)
	}

	return s.repo.GetHolidayByID(ctx, nil, created.ID)
}

func (s *holidayService) UpdateHoliday(ctx context.Context, id uint, req dto.UpdateHolidayRequest) (dto.HolidayResponse, error) {
	existing, err := s.repo.GetHolidayByID(ctx, nil, id)
	if err != nil {
		return dto.HolidayResponse{}, fmt.Errorf("update holiday: get existing: %w", err)
	}

	// Build update model from existing values
	date, _ := time.Parse("2006-01-02", existing.Date)
	updateModel := model.Holiday{
		Name:        existing.Name,
		Year:        existing.Year,
		Date:        date,
		Type:        model.HolidayTypeEnum(existing.Type),
		BranchID:    existing.BranchID,
		Description: existing.Description,
	}

	if req.Name != nil {
		updateModel.Name = *req.Name
	}
	if req.Date != nil {
		d, err := time.Parse("2006-01-02", *req.Date)
		if err != nil {
			return dto.HolidayResponse{}, fmt.Errorf("invalid date format: %w", err)
		}
		updateModel.Date = d
		updateModel.Year = d.Year()
	}
	if req.Type != nil {
		if !validHolidayTypes[*req.Type] {
			return dto.HolidayResponse{}, fmt.Errorf("invalid holiday type: %q", *req.Type)
		}
		updateModel.Type = model.HolidayTypeEnum(*req.Type)
	}
	if req.BranchID != nil {
		updateModel.BranchID = req.BranchID
	}
	if req.Description != nil {
		updateModel.Description = req.Description
	}

	if _, err := s.repo.UpdateHoliday(ctx, nil, id, updateModel); err != nil {
		return dto.HolidayResponse{}, fmt.Errorf("update holiday: %w", err)
	}

	return s.repo.GetHolidayByID(ctx, nil, id)
}

func (s *holidayService) DeleteHoliday(ctx context.Context, id uint) error {
	if err := s.repo.DeleteHoliday(ctx, nil, id); err != nil {
		return fmt.Errorf("delete holiday: %w", err)
	}
	return nil
}
