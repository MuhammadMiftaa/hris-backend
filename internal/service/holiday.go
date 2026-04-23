package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"hris-backend/config/env"
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
	SyncFromExternalAPI(ctx context.Context, req dto.SyncHolidayRequest) (dto.SyncHolidayResponse, error)
}

type holidayService struct {
	repo repository.HolidayRepository
}

func NewHolidayService(repo repository.HolidayRepository) HolidayService {
	return &holidayService{repo: repo}
}

var validHolidayTypes = map[string]bool{
	"national":   true,
	"joint":      true,
	"observance": true,
	"company":    true,
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

func (s *holidayService) SyncFromExternalAPI(ctx context.Context, req dto.SyncHolidayRequest) (dto.SyncHolidayResponse, error) {
	if req.Year <= 0 {
		req.Year = time.Now().Year()
	}

	// Fetch semua page (size fixed 100 per page per spec)
	var allItems []dto.ExternalHolidayItem
	page := 1
	for {
		items, totalPage, err := s.fetchExternalPage(ctx, req.Year, page)
		if err != nil {
			return dto.SyncHolidayResponse{}, err
		}
		allItems = append(allItems, items...)
		if page >= totalPage {
			break
		}
		page++
	}

	// Map ke model Holiday
	var holidays []model.Holiday
	var errs []string
	for _, item := range allItems {
		parsedDate, err := time.Parse("2006-01-02", item.Date)
		if err != nil {
			errs = append(errs, fmt.Sprintf("skip %s: invalid date format", item.Date))
			continue
		}

		holidayType := s.mapExternalType(item.Type)

		holidays = append(holidays, model.Holiday{
			Name:     item.Name,
			Year:     req.Year,
			Date:     parsedDate,
			Type:     holidayType,
			BranchID: req.BranchID,
		})
	}

	synced, skipped, err := s.repo.UpsertHolidays(ctx, nil, holidays)
	if err != nil {
		return dto.SyncHolidayResponse{}, fmt.Errorf("upsert holidays: %w", err)
	}

	return dto.SyncHolidayResponse{
		Synced:  synced,
		Skipped: skipped,
		Year:    req.Year,
		Errors:  errs,
	}, nil
}

func (s *holidayService) fetchExternalPage(ctx context.Context, year, page int) ([]dto.ExternalHolidayItem, int, error) {
	apiURL := fmt.Sprintf("%s/holidays/indonesia/?year=%d&page=%d",
		env.Cfg.ExternalAPI.IndonesiaHolidayAPIURL, year, page)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("build request: %w", err)
	}
	// Header auth sesuai spec: x-api-co-id
	httpReq.Header.Set("x-api-co-id", env.Cfg.ExternalAPI.IndonesiaHolidayAPIKey)
	httpReq.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, 0, fmt.Errorf("call external API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, 0, fmt.Errorf("invalid or missing API key (x-api-co-id)")
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, 0, fmt.Errorf("rate limit exceeded on external API")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("external API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("read response body: %w", err)
	}

	var apiResp dto.ExternalHolidayAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, 0, fmt.Errorf("parse response JSON: %w", err)
	}
	if !apiResp.IsSuccess {
		return nil, 0, fmt.Errorf("external API error: %s", apiResp.Message)
	}

	totalPage := apiResp.Paging.TotalPage
	if totalPage == 0 {
		totalPage = 1
	}

	return apiResp.Data, totalPage, nil
}

func (s *holidayService) mapExternalType(externalType string) model.HolidayTypeEnum {
	switch externalType {
	case "Public Holiday":
		return model.HolidayNational
	case "National Holiday":
		return model.HolidayNational
	case "Joint Holiday":
		return model.HolidayJoint
	case "Observance":
		return model.HolidayObservance
	default:
		return model.HolidayNational
	}
}
