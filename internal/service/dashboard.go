package service

import (
	"context"
	"log"
	"time"

	"hris-backend/internal/repository"
	"hris-backend/internal/struct/dto"
	"hris-backend/internal/utils"
)

type DashboardService interface {
	GetEmployeeDashboard(ctx context.Context, accountID uint, isTrainer bool) (dto.EmployeeDashboardResponse, error)
	GetHRDDashboard(ctx context.Context, hrID uint) (dto.HRDDashboardResponse, error)
	GetRankings(ctx context.Context) (dto.DashboardRankingsResponse, error)
	GetDashboardMetadata(ctx context.Context, employeeID uint) (dto.DashboardMetadataResponse, error)
}

type dashboardService struct {
	dashboardRepo repository.DashboardRepository
	attendRepo    repository.AttendanceRepository
	mutabaahRepo  repository.MutabaahRepository
}

func NewDashboardService(
	dashboardRepo repository.DashboardRepository,
	attendRepo repository.AttendanceRepository,
	mutabaahRepo repository.MutabaahRepository,
) DashboardService {
	return &dashboardService{
		dashboardRepo: dashboardRepo,
		attendRepo:    attendRepo,
		mutabaahRepo:  mutabaahRepo,
	}
}

func (s *dashboardService) GetEmployeeDashboard(ctx context.Context, employeeID uint, isTrainer bool) (dto.EmployeeDashboardResponse, error) {
	today := utils.TodayDate()
	now := time.Now()
	year, month, _ := now.Date()

	// 1. Today attendance status
	todayStatus, _ := s.dashboardRepo.GetTodayAttendanceStatus(ctx, employeeID, today)

	// 2. Today mutabaah status
	mutabaahTodayStatus := s.buildMutabaahTodayStatus(ctx, employeeID, isTrainer, today)

	// 3. Monthly summary
	monthlySummary, _ := s.dashboardRepo.GetMonthlyAttendanceSummary(ctx, employeeID, year, int(month))

	// 4. Leave balances
	leaveBalances, _ := s.dashboardRepo.GetLeaveBalanceSummary(ctx, employeeID, year)

	// 5. Pending requests
	pendingRequests, _ := s.dashboardRepo.GetPendingRequests(ctx, employeeID)

	if leaveBalances == nil {
		leaveBalances = []dto.LeaveBalanceSummaryDTO{}
	}
	if pendingRequests == nil {
		pendingRequests = []dto.PendingRequestDTO{}
	}

	return dto.EmployeeDashboardResponse{
		Today:           todayStatus,
		MutabaahToday:   mutabaahTodayStatus,
		MonthlySummary:  monthlySummary,
		LeaveBalances:   leaveBalances,
		PendingRequests: pendingRequests,
	}, nil
}

func (s *dashboardService) buildMutabaahTodayStatus(ctx context.Context, employeeID uint, isTrainer bool, today string) *dto.MutabaahTodayStatus {
	mutabaahLog, err := s.mutabaahRepo.GetTodayLog(ctx, nil, employeeID, today)
	if err != nil {
		return &dto.MutabaahTodayStatus{
			HasRecord:   false,
			IsSubmitted: false,
		}
	}

	if mutabaahLog == nil {
		targetPages := 0
		if isTrainer {
			targetPages = 10
		} else {
			targetPages = 5
		}
		attendLog, err := s.attendRepo.GetTodayLog(ctx, nil, employeeID, today)
		if err != nil {
			return nil
		}
		status := &dto.MutabaahTodayStatus{
			HasRecord:   false,
			IsSubmitted: false,
			TargetPages: targetPages,
		}
		if attendLog != nil {
			status.AttendanceLogID = &attendLog.ID
		}
		return status
	}

	var submittedAt *string
	if mutabaahLog.SubmittedAt != nil {
		formatted := mutabaahLog.SubmittedAt.Format("2006-01-02T15:04:05Z")
		submittedAt = &formatted
	}

	mutabaahLogID := mutabaahLog.ID
	attendLogID := mutabaahLog.AttendanceLogID

	return &dto.MutabaahTodayStatus{
		HasRecord:       true,
		IsSubmitted:     mutabaahLog.IsSubmitted,
		SubmittedAt:     submittedAt,
		TargetPages:     mutabaahLog.TargetPages,
		MutabaahLogID:   &mutabaahLogID,
		AttendanceLogID: &attendLogID,
	}
}

func (s *dashboardService) GetHRDDashboard(ctx context.Context, hrID uint) (dto.HRDDashboardResponse, error) {
	today := utils.TodayDate()

	queue, _ := s.dashboardRepo.GetApprovalQueue(ctx, hrID)
	counts, _ := s.dashboardRepo.GetApprovalCounts(ctx, hrID)
	teamAttend, _ := s.dashboardRepo.GetTeamAttendanceSummary(ctx, today)
	teamMutabaah, _ := s.dashboardRepo.GetTeamMutabaahSummary(ctx, today)
	notClockedIn, _ := s.dashboardRepo.GetNotClockedIn(ctx, today)
	expiring, _ := s.dashboardRepo.GetExpiringContracts(ctx, 30)

	if queue == nil {
		queue = []dto.ApprovalQueueItemDTO{}
	}
	if notClockedIn == nil {
		notClockedIn = []dto.NotClockedInDTO{}
	}
	if expiring == nil {
		expiring = []dto.ExpiringContractDTO{}
	}

	return dto.HRDDashboardResponse{
		ApprovalQueue:     queue,
		ApprovalCounts:    counts,
		TeamAttendance:    teamAttend,
		TeamMutabaah:      teamMutabaah,
		NotClockedIn:      notClockedIn,
		ExpiringContracts: expiring,
	}, nil
}

func (s *dashboardService) GetRankings(ctx context.Context) (dto.DashboardRankingsResponse, error) {
	now := time.Now()
	year, month, _ := now.Date()

	fastest, err := s.dashboardRepo.GetFastestArrivalRanking(ctx, year, int(month), 5)
	if err != nil {
		log.Printf("[WARN] GetFastestArrivalRanking failed: %v", err)
	}
	tilawah, err := s.dashboardRepo.GetTopTilawahByDepartment(ctx, year, int(month), 5)
	if err != nil {
		log.Printf("[WARN] GetTopTilawahByDepartment failed: %v", err)
	}
	mostLate, err := s.dashboardRepo.GetMostLateRanking(ctx, year, int(month), 5)
	if err != nil {
		log.Printf("[WARN] GetMostLateRanking failed: %v", err)
	}

	if fastest == nil {
		fastest = []dto.RankingEntryDTO{}
	}
	if tilawah == nil {
		tilawah = []dto.DepartmentRankingDTO{}
	}
	if mostLate == nil {
		mostLate = []dto.RankingEntryDTO{}
	}

	return dto.DashboardRankingsResponse{
		FastestArrival: fastest,
		TopTilawah:     tilawah,
		MostLate:       mostLate,
	}, nil
}

func (s *dashboardService) GetDashboardMetadata(ctx context.Context, employeeID uint) (dto.DashboardMetadataResponse, error) {
	leaveTypeMeta, err := s.dashboardRepo.GetLeaveTypeMeta(ctx)
	if err != nil {
		log.Printf("[WARN] GetLeaveTypeMeta failed: %v", err)
	}
	recentAttendanceMeta, err := s.dashboardRepo.GetRecentAttendanceMeta(ctx, employeeID)
	if err != nil {
		log.Printf("[WARN] GetRecentAttendanceMeta failed: %v", err)
	}

	if leaveTypeMeta == nil {
		leaveTypeMeta = []dto.Meta{}
	}
	if recentAttendanceMeta == nil {
		recentAttendanceMeta = []dto.Meta{}
	}

	return dto.DashboardMetadataResponse{
		LeaveTypeMeta:        leaveTypeMeta,
		RecentAttendanceMeta: recentAttendanceMeta,
	}, nil
}
