package repository

import (
	"context"
	"errors"
	"fmt"

	"hris-backend/internal/struct/dto"
	"hris-backend/internal/struct/model"
	"hris-backend/internal/utils"

	"gorm.io/gorm"
)

type AttendanceRepository interface {
	// Attendance log
	GetTodayLog(ctx context.Context, tx Transaction, employeeID uint, date string) (*dto.AttendanceLogResponse, error)
	GetLogByID(ctx context.Context, tx Transaction, id uint) (*dto.AttendanceLogResponse, error)
	GetAllLogs(ctx context.Context, tx Transaction, params dto.AttendanceListParams) (dto.PaginatedResponse[dto.AttendanceLogResponse], error)
	CreateLog(ctx context.Context, tx Transaction, m model.AttendanceLog) (model.AttendanceLog, error)
	UpdateLog(ctx context.Context, tx Transaction, id uint, updates map[string]interface{}) error

	// Shift context — shift aktif untuk pegawai di tanggal tertentu
	GetActiveSchedule(ctx context.Context, tx Transaction, employeeID uint, date string) (*dto.ShiftDayContext, error)

	// Holiday check
	IsHoliday(ctx context.Context, tx Transaction, branchID *uint, date string) (bool, string, error)

	// Leave check — apakah ada cuti approved untuk hari ini
	GetApprovedLeave(ctx context.Context, tx Transaction, employeeID uint, date string) (*uint, error)

	// Business trip check
	GetApprovedBusinessTrip(ctx context.Context, tx Transaction, employeeID uint, date string) (*uint, error)

	// Permission check — izin terlambat / pulang cepat untuk hari ini yang approved
	GetApprovedPermission(ctx context.Context, tx Transaction, employeeID uint, date string, permType string) (*model.PermissionRequest, error)

	// Overtime check — apakah ada lembur approved untuk hari ini
	GetApprovedOvertime(ctx context.Context, tx Transaction, employeeID uint, date string) (bool, error)

	// Branch — untuk validasi GPS radius
	GetBranchByID(ctx context.Context, tx Transaction, branchID uint) (*model.Branch, error)

	// Employee branch
	GetEmployeeBranchID(ctx context.Context, tx Transaction, employeeID uint) (*uint, error)

	// Cron: ambil semua pegawai yang punya jadwal aktif di tanggal tertentu tapi belum ada log
	GetEmployeesWithActiveScheduleWithoutLog(ctx context.Context, tx Transaction, date string) ([]uint, error)

	// Cron: bulk create absent logs
	BulkCreateAbsentLogs(ctx context.Context, tx Transaction, logs []model.AttendanceLog) error

	// LinkOvertimeToLog — asosiasikan overtime_request approved ke attendance log
	LinkOvertimeToLog(ctx context.Context, tx Transaction, employeeID uint, date string, logID uint) error

	// Override
	GetAllOverrides(ctx context.Context, tx Transaction, params dto.OverrideListParams) (dto.PaginatedResponse[dto.AttendanceOverrideResponse], error)
	GetOverrideByID(ctx context.Context, tx Transaction, id uint) (*dto.AttendanceOverrideResponse, error)
	CreateOverride(ctx context.Context, tx Transaction, m model.AttendanceOverride) (model.AttendanceOverride, error)
	UpdateOverrideStatus(ctx context.Context, tx Transaction, id uint, updates map[string]interface{}) error

	// Manual attendance
	CreateManualAttendance(ctx context.Context, tx Transaction, m model.AttendanceLog) (model.AttendanceLog, error)

	// Metadata
	GetEmployeeMetaList(ctx context.Context, tx Transaction) ([]dto.Meta, error)
	GetBranchMetaList(ctx context.Context, tx Transaction) ([]dto.Meta, error)
	GetDepartmentMetaList(ctx context.Context, tx Transaction) ([]dto.Meta, error)
	GetAttendanceMeta(ctx context.Context, tx Transaction, employeeID *uint) ([]dto.Meta, error)

	GetEmployeeRoleLevel(ctx context.Context, tx Transaction, employeeID uint) (string, error)
}

type attendanceRepository struct {
	db *gorm.DB
}

func NewAttendanceRepository(db *gorm.DB) AttendanceRepository {
	return &attendanceRepository{db: db}
}

func (r *attendanceRepository) getDB(ctx context.Context, tx Transaction) (*gorm.DB, error) {
	if tx != nil {
		gormTx, ok := tx.(*GormTx)
		if !ok {
			return nil, errors.New("invalid transaction type")
		}
		return gormTx.db.WithContext(ctx), nil
	}
	return r.db.WithContext(ctx), nil
}

func (r *attendanceRepository) GetTodayLog(ctx context.Context, tx Transaction, employeeID uint, date string) (*dto.AttendanceLogResponse, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}

	var log dto.AttendanceLogResponse
	err = db.Raw(`
		SELECT
			al.id,
			al.employee_id,
			e.full_name              AS employee_name,
			al.attendance_date::TEXT AS attendance_date,
			al.schedule_id,
			st.name                  AS shift_name,
			al.clock_in_at,
			al.clock_out_at,
			al.clock_in_photo_url,
			al.clock_out_photo_url,
			al.clock_in_method::TEXT  AS clock_in_method,
			al.clock_out_method::TEXT AS clock_out_method,
			al.status::TEXT           AS status,
			al.late_minutes,
			al.early_leave_minutes,
			al.overtime_minutes,
			al.is_counted_as_full_day,
			al.permission_request_id,
			al.leave_request_id,
			al.business_trip_request_id,
			al.created_at,
			al.updated_at
		FROM attendance_logs al
		JOIN employees e ON e.id = al.employee_id
		LEFT JOIN employee_schedules es ON es.id = al.schedule_id AND es.deleted_at IS NULL
		LEFT JOIN shift_templates st ON st.id = es.shift_template_id AND st.deleted_at IS NULL
		WHERE al.employee_id = ? AND al.attendance_date = ? AND al.deleted_at IS NULL
	`, employeeID, date).Scan(&log).Error
	if err != nil {
		return nil, err
	}
	if log.ID == 0 {
		return nil, nil
	}
	return &log, nil
}

func (r *attendanceRepository) GetLogByID(ctx context.Context, tx Transaction, id uint) (*dto.AttendanceLogResponse, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}

	var log dto.AttendanceLogResponse
	err = db.Raw(`
		SELECT
			al.id,
			al.employee_id,
			e.full_name              AS employee_name,
			al.attendance_date::TEXT AS attendance_date,
			al.schedule_id,
			st.name                  AS shift_name,
			al.clock_in_at,
			al.clock_out_at,
			al.clock_in_photo_url,
			al.clock_out_photo_url,
			al.clock_in_method::TEXT  AS clock_in_method,
			al.clock_out_method::TEXT AS clock_out_method,
			al.status::TEXT           AS status,
			al.late_minutes,
			al.early_leave_minutes,
			al.overtime_minutes,
			al.is_counted_as_full_day,
			al.permission_request_id,
			al.leave_request_id,
			al.business_trip_request_id,
			al.created_at,
			al.updated_at
		FROM attendance_logs al
		JOIN employees e ON e.id = al.employee_id
		LEFT JOIN employee_schedules es ON es.id = al.schedule_id AND es.deleted_at IS NULL
		LEFT JOIN shift_templates st ON st.id = es.shift_template_id AND st.deleted_at IS NULL
		WHERE al.id = ? AND al.deleted_at IS NULL
	`, id).Scan(&log).Error
	if err != nil {
		return nil, err
	}
	if log.ID == 0 {
		return nil, fmt.Errorf("attendance log not found")
	}
	return &log, nil
}

func (r *attendanceRepository) GetAllLogs(ctx context.Context, tx Transaction, params dto.AttendanceListParams) (dto.PaginatedResponse[dto.AttendanceLogResponse], error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return dto.PaginatedResponse[dto.AttendanceLogResponse]{}, err
	}

	baseQuery := `
		FROM attendance_logs al
		JOIN employees e ON e.id = al.employee_id
		LEFT JOIN departments d ON d.id = e.department_id AND d.deleted_at IS NULL
		LEFT JOIN employee_schedules es ON es.id = al.schedule_id AND es.deleted_at IS NULL
		LEFT JOIN shift_templates st ON st.id = es.shift_template_id AND st.deleted_at IS NULL
		WHERE al.deleted_at IS NULL
	`
	args := []interface{}{}

	if params.Employee != nil && *params.Employee != "" {
		baseQuery += " AND e.full_name ILIKE ?"
		like := "%" + *params.Employee + "%"
		args = append(args, like)
	}
	if params.DepartmentID != nil {
		baseQuery += " AND e.department_id = ?"
		args = append(args, *params.DepartmentID)
	}
	if params.StartDate != nil {
		baseQuery += " AND al.attendance_date >= ?"
		args = append(args, *params.StartDate)
	}
	if params.EndDate != nil {
		baseQuery += " AND al.attendance_date <= ?"
		args = append(args, *params.EndDate)
	}
	if params.Status != nil {
		baseQuery += " AND al.status = ?"
		args = append(args, *params.Status)
	}
	if params.BranchID != nil {
		baseQuery += " AND e.branch_id = ?"
		args = append(args, *params.BranchID)
	}
	if params.Method != nil && *params.Method != "" {
		baseQuery += " AND al.clock_in_method = ?::clock_method_enum"
		args = append(args, *params.Method)
	}

	var total int
	if err := db.Raw("SELECT COUNT(*) "+baseQuery, args...).Scan(&total).Error; err != nil {
		return dto.PaginatedResponse[dto.AttendanceLogResponse]{}, err
	}

	selectQuery := `SELECT
			al.id,
			al.employee_id,
			e.full_name              AS employee_name,
			e.employee_number        AS employee_number,
			d.name                   AS department_name,
			al.attendance_date::TEXT AS attendance_date,
			al.schedule_id,
			st.name                  AS shift_name,
			al.clock_in_at,
			al.clock_out_at,
			al.clock_in_photo_url,
			al.clock_out_photo_url,
			al.clock_in_method::TEXT  AS clock_in_method,
			al.clock_out_method::TEXT AS clock_out_method,
			al.clock_in_lat,
			al.clock_in_lng,
			al.clock_out_lat,
			al.clock_out_lng,
			al.status::TEXT           AS status,
			al.late_minutes,
			al.early_leave_minutes,
			al.overtime_minutes,
			al.is_counted_as_full_day,
			al.permission_request_id,
			al.leave_request_id,
			al.business_trip_request_id,
			al.created_at,
			al.updated_at
		` + baseQuery

	selectQuery += utils.BuildSortClause("attendance", params.SortBy, params.GetSortDir(), "al.attendance_date DESC")
	selectQuery += utils.BuildPaginationClause(params.PaginationParams)

	var logs []dto.AttendanceLogResponse
	if err := db.Raw(selectQuery, args...).Scan(&logs).Error; err != nil {
		return dto.PaginatedResponse[dto.AttendanceLogResponse]{}, err
	}

	perPage := params.GetPerPage()
	totalPage := 1
	if perPage > 0 && total > 0 {
		totalPage = (total + perPage - 1) / perPage
	}

	return dto.PaginatedResponse[dto.AttendanceLogResponse]{
		Data: logs,
		Pagination: dto.PaginationMeta{
			Page:      params.GetPage(),
			PerPage:   perPage,
			Total:     total,
			TotalPage: totalPage,
		},
	}, nil
}

func (r *attendanceRepository) CreateLog(ctx context.Context, tx Transaction, m model.AttendanceLog) (model.AttendanceLog, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return model.AttendanceLog{}, err
	}
	if err := db.Create(&m).Error; err != nil {
		return model.AttendanceLog{}, err
	}
	return m, nil
}

func (r *attendanceRepository) UpdateLog(ctx context.Context, tx Transaction, id uint, updates map[string]interface{}) error {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return err
	}
	return db.Model(&model.AttendanceLog{}).Where("id = ?", id).Updates(updates).Error
}

func (r *attendanceRepository) GetActiveSchedule(ctx context.Context, tx Transaction, employeeID uint, date string) (*dto.ShiftDayContext, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}

	var ctx2 struct {
		ScheduleID      uint    `db:"schedule_id"`
		ShiftTemplateID uint    `db:"shift_template_id"`
		ShiftName       string  `db:"shift_name"`
		IsFlexible      bool    `db:"is_flexible"`
		CanWFA          bool    `db:"can_wfa"`
		DayOfWeek       string  `db:"day_of_week"`
		IsWorkingDay    bool    `db:"is_working_day"`
		ClockInStart    *string `db:"clock_in_start"`
		ClockInEnd      *string `db:"clock_in_end"`
		BreakDhuhrStart *string `db:"break_dhuhr_start"`
		BreakDhuhrEnd   *string `db:"break_dhuhr_end"`
		BreakAsrStart   *string `db:"break_asr_start"`
		BreakAsrEnd     *string `db:"break_asr_end"`
		ClockOutStart   *string `db:"clock_out_start"`
		ClockOutEnd     *string `db:"clock_out_end"`
	}

	err = db.Raw(`
		SELECT
			es.id                              AS schedule_id,
			st.id                              AS shift_template_id,
			st.name                            AS shift_name,
			st.is_flexible,
			st.can_wfa,
			LOWER(TRIM(TO_CHAR($2::DATE, 'Day'))) AS day_of_week,
			COALESCE(std.is_working_day, TRUE) AS is_working_day,
			std.clock_in_start::TEXT           AS clock_in_start,
			std.clock_in_end::TEXT             AS clock_in_end,
			std.break_dhuhr_start::TEXT        AS break_dhuhr_start,
			std.break_dhuhr_end::TEXT          AS break_dhuhr_end,
			std.break_asr_start::TEXT          AS break_asr_start,
			std.break_asr_end::TEXT            AS break_asr_end,
			std.clock_out_start::TEXT          AS clock_out_start,
			std.clock_out_end::TEXT            AS clock_out_end
		FROM employee_schedules es
		JOIN shift_templates st ON st.id = es.shift_template_id AND st.deleted_at IS NULL
		LEFT JOIN shift_template_details std
			ON std.shift_template_id = st.id
			AND std.day_of_week = LOWER(TRIM(TO_CHAR($2::DATE, 'Day')))::day_of_week_enum
			AND std.deleted_at IS NULL
		WHERE es.employee_id = $1
		  AND es.effective_date <= $2::DATE
		  AND (es.end_date IS NULL OR es.end_date >= $2::DATE)
		  AND es.is_active = TRUE
		  AND es.deleted_at IS NULL
		ORDER BY es.effective_date DESC
		LIMIT 1
	`, employeeID, date).Scan(&ctx2).Error
	if err != nil {
		return nil, err
	}
	if ctx2.ScheduleID == 0 {
		return nil, nil // no active schedule
	}

	return &dto.ShiftDayContext{
		ScheduleID:      ctx2.ScheduleID,
		ShiftTemplateID: ctx2.ShiftTemplateID,
		ShiftName:       ctx2.ShiftName,
		IsFlexible:      ctx2.IsFlexible,
		CanWFA:          ctx2.CanWFA,
		DayOfWeek:       ctx2.DayOfWeek,
		IsWorkingDay:    ctx2.IsWorkingDay,
		ClockInStart:    ctx2.ClockInStart,
		ClockInEnd:      ctx2.ClockInEnd,
		BreakDhuhrStart: ctx2.BreakDhuhrStart,
		BreakDhuhrEnd:   ctx2.BreakDhuhrEnd,
		BreakAsrStart:   ctx2.BreakAsrStart,
		BreakAsrEnd:     ctx2.BreakAsrEnd,
		ClockOutStart:   ctx2.ClockOutStart,
		ClockOutEnd:     ctx2.ClockOutEnd,
	}, nil
}

func (r *attendanceRepository) IsHoliday(ctx context.Context, tx Transaction, branchID *uint, date string) (bool, string, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return false, "", err
	}

	var holiday struct {
		Name string `db:"name"`
	}

	// cek libur nasional (branch_id IS NULL) atau libur cabang
	query := `
		SELECT name FROM holidays
		WHERE date = $1
		  AND deleted_at IS NULL
		  AND (branch_id IS NULL`
	args := []interface{}{date}

	if branchID != nil {
		query += " OR branch_id = $2"
		args = append(args, *branchID)
	}
	query += ") LIMIT 1"

	if err := db.Raw(query, args...).Scan(&holiday).Error; err != nil {
		return false, "", err
	}
	if holiday.Name == "" {
		return false, "", nil
	}
	return true, holiday.Name, nil
}

func (r *attendanceRepository) GetApprovedLeave(ctx context.Context, tx Transaction, employeeID uint, date string) (*uint, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}

	var id uint
	err = db.Raw(`
		SELECT id FROM leave_requests
		WHERE employee_id = ?
		  AND status IN ('approved_hr', 'approved_leader')
		  AND start_date <= ?::DATE
		  AND end_date >= ?::DATE
		  AND deleted_at IS NULL
		LIMIT 1
	`, employeeID, date, date).Scan(&id).Error
	if err != nil {
		return nil, err
	}
	if id == 0 {
		return nil, nil
	}
	return &id, nil
}

func (r *attendanceRepository) GetApprovedBusinessTrip(ctx context.Context, tx Transaction, employeeID uint, date string) (*uint, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}

	var id uint
	err = db.Raw(`
		SELECT id FROM business_trip_requests
		WHERE employee_id = ?
		  AND status = 'approved'
		  AND start_date <= ?::DATE
		  AND end_date >= ?::DATE
		  AND deleted_at IS NULL
		LIMIT 1
	`, employeeID, date, date).Scan(&id).Error
	if err != nil {
		return nil, err
	}
	if id == 0 {
		return nil, nil
	}
	return &id, nil
}

func (r *attendanceRepository) GetApprovedPermission(ctx context.Context, tx Transaction, employeeID uint, date string, permType string) (*model.PermissionRequest, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}

	var perm model.PermissionRequest
	err = db.Raw(`
		SELECT * FROM permission_requests
		WHERE employee_id = ?
		  AND date = ?::DATE
		  AND permission_type = ?::permission_type_enum
		  AND status = 'approved'
		  AND deleted_at IS NULL
		LIMIT 1
	`, employeeID, date, permType).Scan(&perm).Error
	if err != nil {
		return nil, err
	}
	if perm.ID == 0 {
		return nil, nil
	}
	return &perm, nil
}

func (r *attendanceRepository) GetApprovedOvertime(ctx context.Context, tx Transaction, employeeID uint, date string) (bool, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return false, err
	}

	var count int64
	err = db.Raw(`
		SELECT COUNT(*) FROM overtime_requests
		WHERE employee_id = ?
		  AND overtime_date = ?::DATE
		  AND status = 'approved_hr'
		  AND deleted_at IS NULL
	`, employeeID, date).Scan(&count).Error
	return count > 0, err
}

func (r *attendanceRepository) GetBranchByID(ctx context.Context, tx Transaction, branchID uint) (*model.Branch, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}
	var branch model.Branch
	if err := db.Where("id = ? AND deleted_at IS NULL", branchID).First(&branch).Error; err != nil {
		return nil, err
	}
	return &branch, nil
}

func (r *attendanceRepository) GetEmployeeBranchID(ctx context.Context, tx Transaction, employeeID uint) (*uint, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}
	var branchID *uint
	err = db.Raw(`SELECT branch_id FROM employees WHERE id = ? AND deleted_at IS NULL`, employeeID).Scan(&branchID).Error
	return branchID, err
}

func (r *attendanceRepository) GetEmployeesWithActiveScheduleWithoutLog(ctx context.Context, tx Transaction, date string) ([]uint, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}

	var ids []uint
	// Ambil semua employee yang punya jadwal aktif & hari itu hari kerja
	// tapi belum punya attendance_log
	err = db.Raw(`
		SELECT DISTINCT es.employee_id
		FROM employee_schedules es
		JOIN shift_template_details std
			ON std.shift_template_id = es.shift_template_id
			AND std.day_of_week = LOWER(TRIM(TO_CHAR($1::DATE, 'Day')))::day_of_week_enum
			AND std.is_working_day = TRUE
			AND std.deleted_at IS NULL
		WHERE es.effective_date <= $1::DATE
		  AND (es.end_date IS NULL OR es.end_date >= $1::DATE)
		  AND es.is_active = TRUE
		  AND es.deleted_at IS NULL
		  AND NOT EXISTS (
			  SELECT 1 FROM attendance_logs al
			  WHERE al.employee_id = es.employee_id
			    AND al.attendance_date = $1::DATE
			    AND al.deleted_at IS NULL
		  )
	`, date).Scan(&ids).Error
	return ids, err
}

func (r *attendanceRepository) BulkCreateAbsentLogs(ctx context.Context, tx Transaction, logs []model.AttendanceLog) error {
	if len(logs) == 0 {
		return nil
	}
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return err
	}
	return db.Create(&logs).Error
}

// LinkOvertimeToLog — update overtime_requests.attendance_log_id
// untuk overtime yang sudah approved di hari yang sama (skenario planned overtime)
func (r *attendanceRepository) LinkOvertimeToLog(ctx context.Context, tx Transaction, employeeID uint, date string, logID uint) error {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return err
	}
	return db.Exec(`
		UPDATE overtime_requests
		SET attendance_log_id = ?, updated_at = NOW()
		WHERE employee_id = ?
		  AND overtime_date = ?::DATE
		  AND status = 'approved_hr'
		  AND (attendance_log_id IS NULL OR attendance_log_id = ?)
		  AND deleted_at IS NULL
	`, logID, employeeID, date, logID).Error
}

func (r *attendanceRepository) GetAllOverrides(ctx context.Context, tx Transaction, params dto.OverrideListParams) (dto.PaginatedResponse[dto.AttendanceOverrideResponse], error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return dto.PaginatedResponse[dto.AttendanceOverrideResponse]{}, err
	}

	baseQuery := `
		FROM attendance_overrides ao
		JOIN attendance_logs al ON al.id = ao.attendance_log_id
		JOIN employees e1 ON e1.id = al.employee_id
		LEFT JOIN departments d ON d.id = e1.department_id AND d.deleted_at IS NULL
		LEFT JOIN employees e2 ON e2.id = ao.approved_by
		WHERE ao.deleted_at IS NULL
	`
	args := []interface{}{}

	if params.Employee != nil && *params.Employee != "" {
		baseQuery += " AND e1.full_name ILIKE ?"
		like := "%" + *params.Employee + "%"
		args = append(args, like)
	}
	if params.Status != nil && *params.Status != "" {
		baseQuery += " AND ao.status = ?"
		args = append(args, *params.Status)
	}
	if params.DepartmentID != nil {
		baseQuery += " AND e1.department_id = ?"
		args = append(args, *params.DepartmentID)
	}
	if params.StartDate != nil && *params.StartDate != "" {
		baseQuery += " AND al.attendance_date >= ?::DATE"
		args = append(args, *params.StartDate)
	}
	if params.EndDate != nil && *params.EndDate != "" {
		baseQuery += " AND al.attendance_date <= ?::DATE"
		args = append(args, *params.EndDate)
	}
	if params.OverrideType != nil && *params.OverrideType != "" {
		baseQuery += " AND ao.override_type = ?"
		args = append(args, *params.OverrideType)
	}

	var total int
	if err := db.Raw("SELECT COUNT(*) "+baseQuery, args...).Scan(&total).Error; err != nil {
		return dto.PaginatedResponse[dto.AttendanceOverrideResponse]{}, err
	}

	selectQuery := `
		SELECT
			ao.id,
			ao.attendance_log_id,
			al.attendance_date::TEXT AS attendance_date,
			ao.requested_by,
			e1.full_name             AS requester_name,
			d.name                   AS department_name,
			ao.approved_by,
			e2.full_name             AS approver_name,
			ao.override_type,
			ao.original_clock_in,
			ao.original_clock_out,
			ao.corrected_clock_in,
			ao.corrected_clock_out,
			ao.reason,
			ao.status,
			ao.created_at,
			ao.updated_at
	` + baseQuery

	selectQuery += utils.BuildSortClause("attendance_override", params.SortBy, params.GetSortDir(), "ao.created_at DESC")
	selectQuery += utils.BuildPaginationClause(params.PaginationParams)

	var resp []dto.AttendanceOverrideResponse
	if err := db.Raw(selectQuery, args...).Scan(&resp).Error; err != nil {
		return dto.PaginatedResponse[dto.AttendanceOverrideResponse]{}, err
	}

	perPage := params.GetPerPage()
	totalPage := 1
	if perPage > 0 && total > 0 {
		totalPage = (total + perPage - 1) / perPage
	}

	return dto.PaginatedResponse[dto.AttendanceOverrideResponse]{
		Data: resp,
		Pagination: dto.PaginationMeta{
			Page:      params.GetPage(),
			PerPage:   perPage,
			Total:     total,
			TotalPage: totalPage,
		},
	}, nil
}

func (r *attendanceRepository) GetOverrideByID(ctx context.Context, tx Transaction, id uint) (*dto.AttendanceOverrideResponse, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}

	var resp dto.AttendanceOverrideResponse
	err = db.Raw(`
		SELECT
			ao.id,
			ao.attendance_log_id,
			al.attendance_date::TEXT AS attendance_date,
			ao.requested_by,
			e1.full_name             AS requester_name,
			ao.approved_by,
			e2.full_name             AS approver_name,
			ao.override_type,
			ao.original_clock_in,
			ao.original_clock_out,
			ao.corrected_clock_in,
			ao.corrected_clock_out,
			ao.reason,
			ao.status,
			ao.created_at,
			ao.updated_at
		FROM attendance_overrides ao
		JOIN attendance_logs al ON al.id = ao.attendance_log_id
		JOIN employees e1 ON e1.id = al.employee_id
		LEFT JOIN employees e2 ON e2.id = ao.approved_by
		WHERE ao.id = ? AND ao.deleted_at IS NULL
	`, id).Scan(&resp).Error
	if err != nil {
		return nil, err
	}
	if resp.ID == 0 {
		return nil, fmt.Errorf("override not found")
	}
	return &resp, nil
}

func (r *attendanceRepository) CreateOverride(ctx context.Context, tx Transaction, m model.AttendanceOverride) (model.AttendanceOverride, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return m, err
	}
	if err := db.Create(&m).Error; err != nil {
		return m, err
	}
	return m, nil
}

func (r *attendanceRepository) UpdateOverrideStatus(ctx context.Context, tx Transaction, id uint, updates map[string]interface{}) error {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return err
	}
	return db.Model(&model.AttendanceOverride{}).Where("id = ?", id).Updates(updates).Error
}

func (r *attendanceRepository) CreateManualAttendance(ctx context.Context, tx Transaction, m model.AttendanceLog) (model.AttendanceLog, error) {
	return r.CreateLog(ctx, tx, m) // Reuses existing CreateLog
}

func (r *attendanceRepository) GetEmployeeMetaList(ctx context.Context, tx Transaction) ([]dto.Meta, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}
	var meta []dto.Meta
	err = db.Raw(`
		SELECT id::TEXT, full_name AS name
		FROM employees
		WHERE deleted_at IS NULL
		ORDER BY full_name ASC
	`).Scan(&meta).Error
	return meta, err
}

func (r *attendanceRepository) GetBranchMetaList(ctx context.Context, tx Transaction) ([]dto.Meta, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}
	var meta []dto.Meta
	err = db.Raw(`
		SELECT id::TEXT, name
		FROM branches
		WHERE deleted_at IS NULL
		ORDER BY name ASC
	`).Scan(&meta).Error
	return meta, err
}

func (r *attendanceRepository) GetDepartmentMetaList(ctx context.Context, tx Transaction) ([]dto.Meta, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}
	var meta []dto.Meta
	err = db.Raw(`
		SELECT id::TEXT, name
		FROM departments
		WHERE deleted_at IS NULL
		ORDER BY name ASC
	`).Scan(&meta).Error
	return meta, err
}

func (r *attendanceRepository) GetAttendanceMeta(ctx context.Context, tx Transaction, employeeID *uint) ([]dto.Meta, error) {
	var meta []dto.Meta
	var query string
	var args []interface{}

	if employeeID != nil {
		query = `
			SELECT 
				id::TEXT AS id,
				TO_CHAR(attendance_date, 'DD Month YYYY') || 
					CASE 
						WHEN clock_in_at IS NOT NULL AND clock_out_at IS NOT NULL 
						THEN ' (' || TO_CHAR(clock_in_at, 'FMHH24:MI') || ' - ' || TO_CHAR(clock_out_at, 'FMHH24:MI') || ')'
						WHEN clock_in_at IS NOT NULL 
						THEN ' (' || TO_CHAR(clock_in_at, 'FMHH24:MI') || ' - ' || CHR(63) || ')'
						ELSE '' 
					END AS name
			FROM attendance_logs
			WHERE employee_id = ? AND deleted_at IS NULL
			ORDER BY attendance_date DESC
			LIMIT 7
		`
		args = []interface{}{*employeeID}
	} else {
		query = `
			SELECT 
				al.id::TEXT AS id,
				TO_CHAR(al.attendance_date, 'DD Month YYYY') || ' | ' || e.full_name || 
					CASE 
						WHEN al.clock_in_at IS NOT NULL AND al.clock_out_at IS NOT NULL 
						THEN ' (' || TO_CHAR(al.clock_in_at, 'FMHH24:MI') || ' - ' || TO_CHAR(al.clock_out_at, 'FMHH24:MI') || ')'
						WHEN al.clock_in_at IS NOT NULL 
						THEN ' (' || TO_CHAR(al.clock_in_at, 'FMHH24:MI') || ' - ' || CHR(63) || ')'
						ELSE '' 
					END AS name
			FROM attendance_logs al
			JOIN employees e ON e.id = al.employee_id
			WHERE al.deleted_at IS NULL
			ORDER BY al.attendance_date DESC
			LIMIT 7
		`
		args = []interface{}{}
	}

	db, err := r.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}

	err = db.Raw(query, args...).Scan(&meta).Error
	return meta, err
}

func (r *attendanceRepository) GetEmployeeRoleLevel(ctx context.Context, tx Transaction, employeeID uint) (string, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return "", err
	}
	var level string
	err = db.Raw(`
		SELECT r.level::TEXT
		FROM employees e
		JOIN accounts a ON a.employee_id = e.id AND a.deleted_at IS NULL
		JOIN roles r ON r.id = a.role_id AND r.deleted_at IS NULL
		WHERE e.id = ? AND e.deleted_at IS NULL
	`, employeeID).Scan(&level).Error
	return level, err
}
