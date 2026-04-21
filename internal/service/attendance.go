package service

import (
	"context"
	"fmt"
	"time"

	"hris-backend/config/log"
	"hris-backend/config/storage"
	"hris-backend/internal/repository"
	"hris-backend/internal/struct/dto"
	"hris-backend/internal/struct/model"
	"hris-backend/internal/utils"
	"hris-backend/internal/utils/data"
)

type AttendanceService interface {
	PresignClockPhoto(ctx context.Context, employeeID uint, action string) (dto.AttendancePresignResponse, error)
	GetPhotoURL(ctx context.Context, objectKey string) (string, error)
	GetTodayStatus(ctx context.Context, employeeID uint) (dto.AttendanceTodayResponse, error)
	ClockIn(ctx context.Context, employeeID uint, req dto.ClockInRequest) (dto.AttendanceLogResponse, error)
	ClockOut(ctx context.Context, employeeID uint, req dto.ClockOutRequest) (dto.AttendanceLogResponse, error)
	GetAllLogs(ctx context.Context, params dto.AttendanceListParams) ([]dto.AttendanceLogResponse, error)
	GetMetadata(ctx context.Context) (dto.AttendanceMetadata, error)
	CreateManualAttendance(ctx context.Context, employeeID uint, req dto.CreateManualAttendanceRequest) (dto.AttendanceLogResponse, error)
	GetAllOverrides(ctx context.Context, params dto.OverrideListParams) ([]dto.AttendanceOverrideResponse, error)
	GetOverrideByID(ctx context.Context, id uint) (dto.AttendanceOverrideResponse, error)
	CreateOverride(ctx context.Context, employeeID uint, req dto.CreateOverrideRequest) (dto.AttendanceOverrideResponse, error)
	UpdateOverrideStatus(ctx context.Context, employeeID uint, id uint, req dto.UpdateOverrideStatusRequest) (dto.AttendanceOverrideResponse, error)
}

type attendanceService struct {
	repo      repository.AttendanceRepository
	txManager repository.TxManager
	minio     storage.MinioClient
}

func NewAttendanceService(
	repo repository.AttendanceRepository,
	txManager repository.TxManager,
	minio storage.MinioClient,
) AttendanceService {
	return &attendanceService{repo: repo, txManager: txManager, minio: minio}
}

func (s *attendanceService) PresignClockPhoto(ctx context.Context, employeeID uint, action string) (dto.AttendancePresignResponse, error) {
	if action != "clock_in" && action != "clock_out" {
		return dto.AttendancePresignResponse{}, fmt.Errorf("action harus 'clock_in' atau 'clock_out'")
	}

	today := utils.TodayDate()
	objectKey := fmt.Sprintf("attendance/%d/%s/%s_%d.jpg",
		employeeID, today, action, time.Now().UnixNano())

	uploadURL, err := s.minio.PresignedPutObject(ctx, storage.BucketAttendance, objectKey, storage.PresignedUploadExpiry)
	if err != nil {
		return dto.AttendancePresignResponse{}, fmt.Errorf("gagal membuat upload URL: %w", err)
	}

	return dto.AttendancePresignResponse{
		UploadURL: uploadURL,
		ObjectKey: objectKey,
		ExpiresIn: int(storage.PresignedUploadExpiry.Seconds()),
	}, nil
}

func (s *attendanceService) GetPhotoURL(ctx context.Context, objectKey string) (string, error) {
	if objectKey == "" {
		return "", fmt.Errorf("object key tidak boleh kosong")
	}
	url, err := s.minio.PresignedGetObject(ctx, storage.BucketAttendance, objectKey, storage.PresignedDownloadExpiry)
	if err != nil {
		return "", fmt.Errorf("gagal membuat download URL: %w", err)
	}
	return url, nil
}

func (s *attendanceService) GetTodayStatus(ctx context.Context, employeeID uint) (dto.AttendanceTodayResponse, error) {
	today := utils.TodayDate()

	log, err := s.repo.GetTodayLog(ctx, nil, employeeID, today)
	if err != nil {
		return dto.AttendanceTodayResponse{}, fmt.Errorf("get today log: %w", err)
	}

	branchID, err := s.repo.GetEmployeeBranchID(ctx, nil, employeeID)
	if err != nil {
		return dto.AttendanceTodayResponse{}, fmt.Errorf("get branch: %w", err)
	}

	isHoliday, holidayName, err := s.repo.IsHoliday(ctx, nil, branchID, today)
	if err != nil {
		return dto.AttendanceTodayResponse{}, fmt.Errorf("check holiday: %w", err)
	}

	shift, err := s.repo.GetActiveSchedule(ctx, nil, employeeID, today)
	if err != nil {
		return dto.AttendanceTodayResponse{}, fmt.Errorf("get schedule: %w", err)
	}

	leaveID, _ := s.repo.GetApprovedLeave(ctx, nil, employeeID, today)
	tripID, _ := s.repo.GetApprovedBusinessTrip(ctx, nil, employeeID, today)
	latePerm, _ := s.repo.GetApprovedPermission(ctx, nil, employeeID, today, "late_arrival")
	earlyPerm, _ := s.repo.GetApprovedPermission(ctx, nil, employeeID, today, "early_leave")

	hasLeave := leaveID != nil
	hasTrip := tripID != nil
	hasPerm := latePerm != nil || earlyPerm != nil
	isWorkingDay := shift != nil && shift.IsWorkingDay

	var holidayNamePtr *string
	if holidayName != "" {
		holidayNamePtr = &holidayName
	}

	canClockIn := false
	canClockOut := false

	if log == nil && !isHoliday && !hasLeave && isWorkingDay {
		canClockIn = true
	}
	if log != nil && log.ClockInAt != nil && log.ClockOutAt == nil {
		canClockOut = true
	}

	resp := dto.AttendanceTodayResponse{
		Log:             log,
		ShiftDetail:     shift,
		IsWorkingDay:    isWorkingDay,
		IsHoliday:       isHoliday,
		HolidayName:     holidayNamePtr,
		CanClockIn:      canClockIn,
		CanClockOut:     canClockOut,
		HasLeave:        hasLeave,
		HasBusinessTrip: hasTrip,
		HasPermission:   hasPerm,
	}
	return resp, nil
}

func (s *attendanceService) ClockIn(ctx context.Context, employeeID uint, req dto.ClockInRequest) (dto.AttendanceLogResponse, error) {
	today := utils.TodayDate()
	now := utils.NowWIB()

	// 1. Cek sudah presensi hari ini
	existing, err := s.repo.GetTodayLog(ctx, nil, employeeID, today)
	if err != nil {
		return dto.AttendanceLogResponse{}, fmt.Errorf("get existing log: %w", err)
	}
	if existing != nil {
		return dto.AttendanceLogResponse{}, fmt.Errorf("sudah melakukan clock in hari ini")
	}

	// 2. Cek hari libur
	branchID, err := s.repo.GetEmployeeBranchID(ctx, nil, employeeID)
	if err != nil {
		return dto.AttendanceLogResponse{}, fmt.Errorf("get branch: %w", err)
	}
	isHoliday, holidayName, err := s.repo.IsHoliday(ctx, nil, branchID, today)
	if err != nil {
		return dto.AttendanceLogResponse{}, fmt.Errorf("check holiday: %w", err)
	}
	if isHoliday {
		return dto.AttendanceLogResponse{}, fmt.Errorf("hari ini adalah hari libur: %s", holidayName)
	}

	// 3. Cek cuti yang disetujui
	leaveID, err := s.repo.GetApprovedLeave(ctx, nil, employeeID, today)
	if err != nil {
		return dto.AttendanceLogResponse{}, fmt.Errorf("check leave: %w", err)
	}
	if leaveID != nil {
		return dto.AttendanceLogResponse{}, fmt.Errorf("anda memiliki cuti yang disetujui untuk hari ini")
	}

	// 4. Ambil jadwal shift aktif
	shift, err := s.repo.GetActiveSchedule(ctx, nil, employeeID, today)
	if err != nil {
		return dto.AttendanceLogResponse{}, fmt.Errorf("get schedule: %w", err)
	}
	if shift == nil {
		return dto.AttendanceLogResponse{}, fmt.Errorf("tidak ada jadwal shift aktif untuk hari ini")
	}
	if !shift.IsWorkingDay {
		return dto.AttendanceLogResponse{}, fmt.Errorf("hari ini bukan hari kerja sesuai jadwal shift")
	}

	// 5. Cek GPS (kecuali ada dinas luar)
	tripID, err := s.repo.GetApprovedBusinessTrip(ctx, nil, employeeID, today)
	if err != nil {
		return dto.AttendanceLogResponse{}, fmt.Errorf("check business trip: %w", err)
	}
	isBusinessTrip := tripID != nil

	clockMethod := model.ClockMethodGPS
	if !isBusinessTrip {
		if branchID == nil {
			return dto.AttendanceLogResponse{}, fmt.Errorf("pegawai tidak memiliki cabang yang terdaftar")
		}
		branch, err := s.repo.GetBranchByID(ctx, nil, *branchID)
		if err != nil {
			return dto.AttendanceLogResponse{}, fmt.Errorf("get branch detail: %w", err)
		}
		if !branch.AllowWFH {
			if branch.Latitude == nil || branch.Longitude == nil {
				return dto.AttendanceLogResponse{}, fmt.Errorf("koordinat cabang belum dikonfigurasi")
			}
			dist := utils.HaversineDistance(req.Latitude, req.Longitude, *branch.Latitude, *branch.Longitude)
			if dist > float64(branch.RadiusMeters) {
				return dto.AttendanceLogResponse{}, fmt.Errorf(
					"bad request: lokasi anda (%.0f meter) di luar radius presensi yang diizinkan (%d meter)",
					dist, branch.RadiusMeters,
				)
			}
		}
	}

	lateMinutes := 0
	status := model.AttendancePresent

	// Cek izin terlambat
	latePerm, err := s.repo.GetApprovedPermission(ctx, nil, employeeID, today, "late_arrival")
	if err != nil {
		return dto.AttendanceLogResponse{}, fmt.Errorf("check late permission: %w", err)
	}

	if !shift.IsFlexible && shift.ClockInEnd != nil {
		clockInEnd, parseErr := utils.ParseTimeString(*shift.ClockInEnd, today)
		if parseErr == nil && now.After(clockInEnd) {
			lateMinutes = int(now.Sub(clockInEnd).Minutes())
			if latePerm != nil {
				// Ada izin terlambat — status tetap present, tapi catat menit
				status = model.AttendancePresent
			} else {
				status = model.AttendanceLate
			}
		}
	}

	if isBusinessTrip {
		status = model.AttendanceBusinessTrip
	}

	// 7. Photo key validasi
	if req.PhotoKey == "" {
		return dto.AttendanceLogResponse{}, fmt.Errorf("foto clock in wajib diisi")
	}

	var permissionRequestID *uint
	if latePerm != nil {
		permissionRequestID = &latePerm.ID
	}

	var businessTripRequestID *uint
	if tripID != nil {
		businessTripRequestID = tripID
	}

	logModel := model.AttendanceLog{
		EmployeeID:            employeeID,
		ScheduleID:            &shift.ScheduleID,
		AttendanceDate:        time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()),
		ClockInAt:             &now,
		ClockInLat:            &req.Latitude,
		ClockInLng:            &req.Longitude,
		ClockInPhotoURL:       &req.PhotoKey,
		ClockInMethod:         &clockMethod,
		Status:                status,
		LateMinutes:           lateMinutes,
		PermissionRequestID:   permissionRequestID,
		BusinessTripRequestID: businessTripRequestID,
		IsAutoGenerated:       false,
	}

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return dto.AttendanceLogResponse{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	created, err := s.repo.CreateLog(ctx, tx, logModel)
	if err != nil {
		return dto.AttendanceLogResponse{}, fmt.Errorf("create log: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return dto.AttendanceLogResponse{}, fmt.Errorf("commit: %w", err)
	}

	resp, err := s.repo.GetLogByID(ctx, nil, created.ID)
	if err != nil || resp == nil {
		return dto.AttendanceLogResponse{}, fmt.Errorf("get created log: %w", err)
	}
	return *resp, nil
}

func (s *attendanceService) ClockOut(ctx context.Context, employeeID uint, req dto.ClockOutRequest) (dto.AttendanceLogResponse, error) {
	today := utils.TodayDate()
	now := utils.NowWIB()

	// 1. Harus sudah clock in dulu
	existing, err := s.repo.GetTodayLog(ctx, nil, employeeID, today)
	if err != nil {
		return dto.AttendanceLogResponse{}, fmt.Errorf("get today log: %w", err)
	}
	if existing == nil {
		return dto.AttendanceLogResponse{}, fmt.Errorf("belum melakukan clock in hari ini")
	}
	if existing.ClockOutAt != nil {
		return dto.AttendanceLogResponse{}, fmt.Errorf("sudah melakukan clock out hari ini")
	}

	shift, err := s.repo.GetActiveSchedule(ctx, nil, employeeID, today)
	if err != nil {
		return dto.AttendanceLogResponse{}, fmt.Errorf("get schedule: %w", err)
	}

	// 2. GPS check (kecuali business trip)
	isBusinessTrip := existing.BusinessTripRequestID != nil

	branchID, err := s.repo.GetEmployeeBranchID(ctx, nil, employeeID)
	if err != nil {
		return dto.AttendanceLogResponse{}, fmt.Errorf("get branch: %w", err)
	}

	clockMethod := model.ClockMethodGPS
	if !isBusinessTrip && branchID != nil {
		branch, err := s.repo.GetBranchByID(ctx, nil, *branchID)
		if err != nil {
			return dto.AttendanceLogResponse{}, fmt.Errorf("get branch detail: %w", err)
		}
		if !branch.AllowWFH && branch.Latitude != nil && branch.Longitude != nil {
			dist := utils.HaversineDistance(req.Latitude, req.Longitude, *branch.Latitude, *branch.Longitude)
			if dist > float64(branch.RadiusMeters) {
				return dto.AttendanceLogResponse{}, fmt.Errorf(
					"lokasi anda (%.0f meter) di luar radius presensi yang diizinkan (%d meter)",
					dist, branch.RadiusMeters,
				)
			}
		}
	}

	// 3. Hitung early leave dan overtime
	earlyLeaveMinutes := 0
	overtimeMinutes := 0
	// Konversi dari string (dto) ke model enum
	newStatus := model.AttendanceStatusEnum(existing.Status)

	// Cek izin pulang cepat
	earlyPerm, err := s.repo.GetApprovedPermission(ctx, nil, employeeID, today, "early_leave")
	if err != nil {
		return dto.AttendanceLogResponse{}, fmt.Errorf("check early leave permission: %w", err)
	}

	if shift != nil && !shift.IsFlexible {
		if shift.ClockOutStart != nil {
			clockOutStart, parseErr := utils.ParseTimeString(*shift.ClockOutStart, today)
			if parseErr == nil && now.Before(clockOutStart) {
				// Pulang lebih awal dari window clock_out_start
				earlyLeaveMinutes = int(clockOutStart.Sub(now).Minutes())
				if earlyPerm == nil {
					// Tidak ada izin → tandai half_day
					newStatus = model.AttendanceHalfDay
				}
				// Jika ada izin → status tidak berubah ke half_day
			}
		}

		if shift.ClockOutEnd != nil {
			clockOutEnd, parseErr := utils.ParseTimeString(*shift.ClockOutEnd, today)
			if parseErr == nil && now.After(clockOutEnd) {
				overtimeMinutes = int(now.Sub(clockOutEnd).Minutes())
			}
		}
	}

	// 4. Photo key validasi
	if req.PhotoKey == "" {
		return dto.AttendanceLogResponse{}, fmt.Errorf("foto clock out wajib diisi")
	}

	updates := map[string]interface{}{
		"clock_out_at":        now,
		"clock_out_lat":       req.Latitude,
		"clock_out_lng":       req.Longitude,
		"clock_out_photo_url": req.PhotoKey,
		"clock_out_method":    clockMethod,
		"early_leave_minutes": earlyLeaveMinutes,
		"overtime_minutes":    overtimeMinutes,
		"status":              newStatus,
		"updated_at":          now,
	}

	// Catat izin pulang cepat jika ada
	if earlyPerm != nil {
		updates["permission_request_id"] = earlyPerm.ID
	}

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return dto.AttendanceLogResponse{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if err := s.repo.UpdateLog(ctx, tx, existing.ID, updates); err != nil {
		return dto.AttendanceLogResponse{}, fmt.Errorf("update log: %w", err)
	}

	// Jika ada overtime, asosiasikan overtime_request ke attendance log ini
	if overtimeMinutes > 0 {
		_ = s.repo.LinkOvertimeToLog(ctx, tx, employeeID, today, existing.ID)
	}

	if err := tx.Commit(); err != nil {
		return dto.AttendanceLogResponse{}, fmt.Errorf("commit: %w", err)
	}

	resp, err := s.repo.GetLogByID(ctx, nil, existing.ID)
	if err != nil || resp == nil {
		return dto.AttendanceLogResponse{}, fmt.Errorf("get updated log: %w", err)
	}
	return *resp, nil
}

func (s *attendanceService) GetAllLogs(ctx context.Context, params dto.AttendanceListParams) ([]dto.AttendanceLogResponse, error) {
	return s.repo.GetAllLogs(ctx, nil, params)
}

func (s *attendanceService) GetMetadata(ctx context.Context) (dto.AttendanceMetadata, error) {
	empMeta, err := s.repo.GetEmployeeMetaList(ctx, nil)
	if err != nil {
		return dto.AttendanceMetadata{}, fmt.Errorf("failed to fetch employee meta: %w", err)
	}

	return dto.AttendanceMetadata{
		StatusMeta:       data.AttendanceStatusMeta,
		ClockMethodMeta:  data.ClockMethodMeta,
		OverrideTypeMeta: data.OverrideTypeMeta,
		EmployeeMeta:     empMeta,
	}, nil
}

func (s *attendanceService) CreateManualAttendance(ctx context.Context, employeeID uint, req dto.CreateManualAttendanceRequest) (dto.AttendanceLogResponse, error) {
	existing, err := s.repo.GetTodayLog(ctx, nil, req.EmployeeID, req.AttendanceDate)
	if err != nil {
		return dto.AttendanceLogResponse{}, fmt.Errorf("check existing: %w", err)
	}
	if existing != nil {
		return dto.AttendanceLogResponse{}, fmt.Errorf("attendance log for %s already exists", req.AttendanceDate)
	}

	shift, err := s.repo.GetActiveSchedule(ctx, nil, req.EmployeeID, req.AttendanceDate)
	if err != nil {
		return dto.AttendanceLogResponse{}, fmt.Errorf("get schedule: %w", err)
	}

	clockMethod := model.ClockMethodManual
	status := model.AttendancePresent
	lateMinutes := 0

	tIn, err := utils.ParseTimeString(req.ClockInAt, req.AttendanceDate)
	if err != nil {
		log.Debug("parse clock in time: %w", map[string]any{"Result": tIn, "clock in": req.ClockInAt, "date": req.AttendanceDate})
		return dto.AttendanceLogResponse{}, fmt.Errorf("parse clock in time: %w", err)
	}

	var tOutPtr *time.Time
	if req.ClockOutAt != nil {
		tOut, err := utils.ParseTimeString(*req.ClockOutAt, req.AttendanceDate)
		if err != nil {
			log.Debug("parse clock out time: %w", map[string]any{"Result": tOut, "clock out": *req.ClockOutAt, "date": req.AttendanceDate})
			return dto.AttendanceLogResponse{}, fmt.Errorf("parse clock out time: %w", err)
		}

		tOutPtr = &tOut
	}

	if shift != nil && !shift.IsFlexible && shift.ClockInEnd != nil {
		clockInEnd, errEnd := utils.ParseTimeString(*shift.ClockInEnd, req.AttendanceDate)
		if errEnd == nil && tIn.After(clockInEnd) {
			status = model.AttendanceLate
			lateMinutes = int(tIn.Sub(clockInEnd).Minutes())
		}
	}

	latePerm, _ := s.repo.GetApprovedPermission(ctx, nil, req.EmployeeID, req.AttendanceDate, "late_arrival")
	var permID *uint
	if latePerm != nil {
		status = model.AttendancePresent
		permID = &latePerm.ID
	}

	logModel := model.AttendanceLog{
		EmployeeID:          req.EmployeeID,
		AttendanceDate:      tIn,
		ClockInAt:           &tIn,
		ClockOutAt:          tOutPtr,
		ClockInMethod:       &clockMethod,
		Status:              status,
		LateMinutes:         lateMinutes,
		IsAutoGenerated:     false,
		PermissionRequestID: permID,
	}

	if tOutPtr != nil {
		logModel.ClockOutMethod = &clockMethod
	}

	if shift != nil {
		logModel.ScheduleID = &shift.ScheduleID
	}

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return dto.AttendanceLogResponse{}, fmt.Errorf("tx begin: %w", err)
	}
	defer tx.Rollback()

	created, err := s.repo.CreateManualAttendance(ctx, tx, logModel)
	if err != nil {
		return dto.AttendanceLogResponse{}, fmt.Errorf("create manual log: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return dto.AttendanceLogResponse{}, fmt.Errorf("commit: %w", err)
	}

	log, err := s.repo.GetLogByID(ctx, nil, created.ID)
	if err != nil || log == nil {
		return dto.AttendanceLogResponse{}, fmt.Errorf("fetch created log: %w", err)
	}
	return *log, nil
}

func (s *attendanceService) GetAllOverrides(ctx context.Context, params dto.OverrideListParams) ([]dto.AttendanceOverrideResponse, error) {
	return s.repo.GetAllOverrides(ctx, nil, params)
}

func (s *attendanceService) GetOverrideByID(ctx context.Context, id uint) (dto.AttendanceOverrideResponse, error) {
	res, err := s.repo.GetOverrideByID(ctx, nil, id)
	if err != nil {
		return dto.AttendanceOverrideResponse{}, err
	}
	return *res, nil
}

func (s *attendanceService) CreateOverride(ctx context.Context, employeeID uint, req dto.CreateOverrideRequest) (dto.AttendanceOverrideResponse, error) {
	log, err := s.repo.GetLogByID(ctx, nil, req.AttendanceLogID)
	if err != nil || log == nil {
		return dto.AttendanceOverrideResponse{}, fmt.Errorf("attendance log not found: %w", err)
	}

	var parsedIn, parsedOut *time.Time
	if req.CorrectedClockIn != nil {
		t, e := utils.ParseTimeString(*req.CorrectedClockIn, log.AttendanceDate)
		if e == nil {
			parsedIn = &t
		}
	}
	if req.CorrectedClockOut != nil {
		t, e := utils.ParseTimeString(*req.CorrectedClockOut, log.AttendanceDate)
		if e == nil {
			parsedOut = &t
		}
	}

	modelType := model.OverrideTypeEnum(req.OverrideType)

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return dto.AttendanceOverrideResponse{}, err
	}
	defer tx.Rollback()

	ov := model.AttendanceOverride{
		AttendanceLogID:   req.AttendanceLogID,
		RequestedBy:       employeeID,
		OverrideType:      modelType,
		OriginalClockIn:   log.ClockInAt,
		OriginalClockOut:  log.ClockOutAt,
		CorrectedClockIn:  parsedIn,
		CorrectedClockOut: parsedOut,
		Reason:            req.Reason,
		Status:            model.RequestStatusPending,
	}

	created, err := s.repo.CreateOverride(ctx, tx, ov)
	if err != nil {
		return dto.AttendanceOverrideResponse{}, fmt.Errorf("create override: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return dto.AttendanceOverrideResponse{}, err
	}

	res, err := s.repo.GetOverrideByID(ctx, nil, created.ID)
	if err != nil {
		return dto.AttendanceOverrideResponse{}, err
	}
	return *res, nil
}

func (s *attendanceService) UpdateOverrideStatus(ctx context.Context, employeeID uint, id uint, req dto.UpdateOverrideStatusRequest) (dto.AttendanceOverrideResponse, error) {
	ov, err := s.repo.GetOverrideByID(ctx, nil, id)
	if err != nil {
		return dto.AttendanceOverrideResponse{}, err
	}
	if ov.Status != string(model.RequestStatusPending) {
		return dto.AttendanceOverrideResponse{}, fmt.Errorf("override is no longer pending")
	}

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return dto.AttendanceOverrideResponse{}, err
	}
	defer tx.Rollback()

	updates := map[string]interface{}{
		"status":      req.Status,
		"approved_by": employeeID,
		"reason":      req.ApproverNotes,
	}

	if err := s.repo.UpdateOverrideStatus(ctx, tx, id, updates); err != nil {
		return dto.AttendanceOverrideResponse{}, err
	}

	if model.RequestStatusEnum(req.Status) == model.RequestStatusApproved {
		logUpds := make(map[string]interface{})
		if ov.CorrectedClockIn != nil {
			logUpds["clock_in_at"] = ov.CorrectedClockIn
		}
		if ov.CorrectedClockOut != nil {
			logUpds["clock_out_at"] = ov.CorrectedClockOut
		}
		logUpds["updated_at"] = time.Now()

		if err := s.repo.UpdateLog(ctx, tx, ov.AttendanceLogID, logUpds); err != nil {
			return dto.AttendanceOverrideResponse{}, fmt.Errorf("syncing corrected values: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return dto.AttendanceOverrideResponse{}, err
	}

	res, err := s.repo.GetOverrideByID(ctx, nil, id)
	if err != nil {
		return dto.AttendanceOverrideResponse{}, err
	}
	return *res, nil
}
