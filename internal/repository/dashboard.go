package repository

import (
	"context"
	"errors"
	"fmt"

	"hris-backend/internal/struct/dto"

	"gorm.io/gorm"
)

type DashboardRepository interface {
	GetTodayAttendanceStatus(ctx context.Context, employeeID uint, date string) (dto.TodayAttendanceStatus, error)
	GetMonthlyAttendanceSummary(ctx context.Context, employeeID uint, year int, month int) (dto.AttendanceSummaryDTO, error)
	GetLeaveBalanceSummary(ctx context.Context, employeeID uint, year int) ([]dto.LeaveBalanceSummaryDTO, error)
	GetEmployeeRequests(ctx context.Context, employeeID uint) ([]dto.EmployeeRequestDTO, error)
	GetApprovalQueue(ctx context.Context, approverID uint) ([]dto.ApprovalQueueItemDTO, error)
	GetApprovalCounts(ctx context.Context, approverID uint) (dto.ApprovalCountsDTO, error)
	GetTeamAttendanceSummary(ctx context.Context, date string) (dto.TeamAttendanceSummaryDTO, error)
	GetTeamMutabaahSummary(ctx context.Context, date string) (dto.TeamMutabaahSummaryDTO, error)
	GetNotClockedIn(ctx context.Context, date string) ([]dto.NotClockedInDTO, error)
	GetExpiringContracts(ctx context.Context, days int) ([]dto.ExpiringContractDTO, error)
	GetFastestArrivalRanking(ctx context.Context, date string, limit int) ([]dto.RankingEntryDTO, error)
	GetTopTilawahByDepartment(ctx context.Context, date string, limit int) ([]dto.DepartmentRankingDTO, error)
	GetFastestMutabaahRanking(ctx context.Context, date string, limit int) ([]dto.RankingEntryDTO, error)
	GetRecentAttendanceMeta(ctx context.Context, employeeID *uint) ([]dto.Meta, error)
	GetLeaveTypeMeta(ctx context.Context) ([]dto.Meta, error)
	GetEmployeeMeta(ctx context.Context, employeeID *uint) ([]dto.Meta, error)
}

type dashboardRepository struct {
	db *gorm.DB
}

func NewDashboardRepository(db *gorm.DB) DashboardRepository {
	return &dashboardRepository{db: db}
}

func (r *dashboardRepository) getDB(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx)
}

// GetTodayAttendanceStatus — bangun TodayAttendanceStatus sesuai kontrak frontend
func (r *dashboardRepository) GetTodayAttendanceStatus(ctx context.Context, employeeID uint, date string) (dto.TodayAttendanceStatus, error) {
	var raw struct {
		ClockInAt   *string `db:"clock_in_at"`
		ClockOutAt  *string `db:"clock_out_at"`
		Status      *string `db:"status"`
		LateMinutes int     `db:"late_minutes"`
	}

	err := r.getDB(ctx).Raw(`
		SELECT
			TO_CHAR(clock_in_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS clock_in_at,
			TO_CHAR(clock_out_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS clock_out_at,
			status::TEXT AS status,
			late_minutes
		FROM attendance_logs
		WHERE employee_id = ? AND attendance_date = ?::DATE AND deleted_at IS NULL
		LIMIT 1
	`, employeeID, date).Scan(&raw).Error
	if err != nil {
		return dto.TodayAttendanceStatus{}, err
	}

	result := dto.TodayAttendanceStatus{
		HasClockedIn:  raw.ClockInAt != nil,
		HasClockedOut: raw.ClockOutAt != nil,
		ClockInAt:     raw.ClockInAt,
		ClockOutAt:    raw.ClockOutAt,
		Status:        raw.Status,
		LateMinutes:   raw.LateMinutes,
	}
	return result, nil
}

func (r *dashboardRepository) GetMonthlyAttendanceSummary(ctx context.Context, employeeID uint, year int, month int) (dto.AttendanceSummaryDTO, error) {
	var summary dto.AttendanceSummaryDTO
	err := r.getDB(ctx).Raw(`
		SELECT
			COUNT(*) FILTER (WHERE status = 'present')       AS total_present,
			COUNT(*) FILTER (WHERE status = 'late')          AS total_late,
			COUNT(*) FILTER (WHERE status = 'absent')        AS total_absent,
			COUNT(*) FILTER (WHERE status = 'leave')         AS total_leave,
			COUNT(*) FILTER (WHERE status = 'business_trip') AS total_business_trip,
			COUNT(*) FILTER (WHERE status = 'half_day')      AS total_half_day
		FROM attendance_logs
		WHERE employee_id = ?
		  AND EXTRACT(YEAR  FROM attendance_date) = ?
		  AND EXTRACT(MONTH FROM attendance_date) = ?
		  AND deleted_at IS NULL
	`, employeeID, year, month).Scan(&summary).Error
	return summary, err
}

func (r *dashboardRepository) GetLeaveBalanceSummary(ctx context.Context, employeeID uint, year int) ([]dto.LeaveBalanceSummaryDTO, error) {
	var summary []dto.LeaveBalanceSummaryDTO
	err := r.getDB(ctx).Raw(`
		SELECT
			lt.id   AS leave_type_id,
			lt.name AS leave_type_name,
			lt.max_occurrences_per_year AS total_quota,
			lb.used_duration            AS used,
			CASE
				WHEN lt.max_total_duration_per_year IS NOT NULL
				THEN lt.max_total_duration_per_year - lb.used_duration
				ELSE NULL
			END AS remaining
		FROM leave_balances lb
		JOIN leave_types lt ON lt.id = lb.leave_type_id
		WHERE lb.employee_id = ? AND lb.year = ? AND lb.deleted_at IS NULL
	`, employeeID, year).Scan(&summary).Error
	return summary, err
}

func (r *dashboardRepository) GetEmployeeRequests(ctx context.Context, employeeID uint) ([]dto.EmployeeRequestDTO, error) {
	var requests []dto.EmployeeRequestDTO
	err := r.getDB(ctx).Raw(`
		SELECT id, 'leave' AS type, 'Cuti' AS label, created_at::TEXT, status::TEXT AS status,
			start_date::TEXT AS date
		FROM leave_requests
		WHERE employee_id = ? AND deleted_at IS NULL
		AND (end_date >= CURRENT_DATE OR created_at::DATE = CURRENT_DATE)
		UNION ALL
		SELECT id, 'permission' AS type, 'Izin' AS label, created_at::TEXT, status::TEXT AS status,
			date::TEXT AS date
		FROM permission_requests
		WHERE employee_id = ? AND deleted_at IS NULL
		AND (date >= CURRENT_DATE OR created_at::DATE = CURRENT_DATE)
		UNION ALL
		SELECT id, 'overtime' AS type, 'Lembur' AS label, created_at::TEXT, status::TEXT AS status,
			overtime_date::TEXT AS date
		FROM overtime_requests
		WHERE employee_id = ? AND deleted_at IS NULL
		AND (overtime_date >= CURRENT_DATE OR created_at::DATE = CURRENT_DATE)
		UNION ALL
		SELECT id, 'business_trip' AS type, 'Tugas' AS label, created_at::TEXT, status::TEXT AS status,
			start_date::TEXT AS date
		FROM business_trip_requests
		WHERE employee_id = ? AND deleted_at IS NULL
		AND (end_date >= CURRENT_DATE OR created_at::DATE = CURRENT_DATE)
		UNION ALL
		SELECT ao.id, 'attendance_override' AS type, 'Koreksi Absen' AS label, ao.created_at::TEXT, ao.status::TEXT AS status,
			al.attendance_date::TEXT AS date
		FROM attendance_overrides ao
		JOIN attendance_logs al ON al.id = ao.attendance_log_id AND al.deleted_at IS NULL
		WHERE ao.requested_by = ? AND ao.deleted_at IS NULL
		AND (al.attendance_date >= CURRENT_DATE OR ao.created_at::DATE = CURRENT_DATE)
		ORDER BY created_at DESC
		LIMIT 10
	`, employeeID, employeeID, employeeID, employeeID, employeeID).Scan(&requests).Error
	return requests, err
}

func (r *dashboardRepository) GetApprovalQueue(ctx context.Context, approverID uint) ([]dto.ApprovalQueueItemDTO, error) {
	var items []dto.ApprovalQueueItemDTO
	err := r.getDB(ctx).Raw(`
		SELECT l.id, 'leave'         AS type, e.full_name AS employee_name, 'Cuti'          AS label, l.created_at::TEXT
		FROM leave_requests l
		JOIN employees e ON e.id = l.employee_id
		WHERE l.status = 'pending' AND l.deleted_at IS NULL
		UNION ALL
		SELECT p.id, 'permission'    AS type, e.full_name AS employee_name, 'Izin'          AS label, p.created_at::TEXT
		FROM permission_requests p
		JOIN employees e ON e.id = p.employee_id
		WHERE p.status = 'pending' AND p.deleted_at IS NULL
		UNION ALL
		SELECT o.id, 'overtime'      AS type, e.full_name AS employee_name, 'Lembur'        AS label, o.created_at::TEXT
		FROM overtime_requests o
		JOIN employees e ON e.id = o.employee_id
		WHERE o.status = 'pending' AND o.deleted_at IS NULL
		UNION ALL
		SELECT b.id, 'business_trip' AS type, e.full_name AS employee_name, 'Tugas'   AS label, b.created_at::TEXT
		FROM business_trip_requests b
		JOIN employees e ON e.id = b.employee_id
		WHERE b.status = 'pending' AND b.deleted_at IS NULL
		UNION ALL
		SELECT a.id, 'override'      AS type, e.full_name AS employee_name, 'Koreksi Absen' AS label, a.created_at::TEXT
		FROM attendance_overrides a
		JOIN employees e ON e.id = a.requested_by
		WHERE a.status = 'pending' AND a.deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT 20
	`).Scan(&items).Error
	return items, err
}

func (r *dashboardRepository) GetApprovalCounts(ctx context.Context, approverID uint) (dto.ApprovalCountsDTO, error) {
	var counts dto.ApprovalCountsDTO
	err := r.getDB(ctx).Raw(`
		SELECT
			(SELECT COUNT(*) FROM leave_requests        WHERE status = 'pending' AND deleted_at IS NULL) AS leave,
			(SELECT COUNT(*) FROM permission_requests   WHERE status = 'pending' AND deleted_at IS NULL) AS permission,
			(SELECT COUNT(*) FROM overtime_requests     WHERE status = 'pending' AND deleted_at IS NULL) AS overtime,
			(SELECT COUNT(*) FROM business_trip_requests WHERE status = 'pending' AND deleted_at IS NULL) AS business_trip,
			(SELECT COUNT(*) FROM attendance_overrides  WHERE status = 'pending' AND deleted_at IS NULL) AS override
	`).Scan(&counts).Error
	if err == nil {
		counts.Total = counts.Leave + counts.Permission + counts.Overtime + counts.BusinessTrip + counts.Override
	}
	return counts, err
}

func (r *dashboardRepository) GetTeamAttendanceSummary(ctx context.Context, date string) (dto.TeamAttendanceSummaryDTO, error) {
	var summary dto.TeamAttendanceSummaryDTO
	err := r.getDB(ctx).Raw(`
		SELECT
			(SELECT COUNT(DISTINCT e.id)
			 FROM employees e
			 INNER JOIN employee_schedules es
				 ON es.employee_id = e.id
				 AND es.is_active = TRUE
				 AND es.effective_date <= ?::DATE
				 AND (es.end_date IS NULL OR es.end_date >= ?::DATE)
				 AND es.deleted_at IS NULL
			 INNER JOIN shift_templates st ON st.id = es.shift_template_id AND st.deleted_at IS NULL
			 INNER JOIN shift_template_details std
				 ON std.shift_template_id = st.id
				 AND std.day_of_week = LOWER(TRIM(TO_CHAR(?::DATE, 'Day')))::day_of_week_enum
				 AND std.is_working_day = TRUE
				 AND std.deleted_at IS NULL
			 WHERE e.deleted_at IS NULL
			) AS total_employees,
			(SELECT COUNT(*) FROM attendance_logs WHERE attendance_date = ?::DATE AND status = 'present'      AND deleted_at IS NULL) AS present_today,
			(SELECT COUNT(*) FROM attendance_logs WHERE attendance_date = ?::DATE AND status = 'late'         AND deleted_at IS NULL) AS late_today,
			(SELECT COUNT(*) FROM attendance_logs WHERE attendance_date = ?::DATE AND status = 'leave'        AND deleted_at IS NULL) AS on_leave
	`, date, date, date, date, date, date).Scan(&summary).Error
	if err != nil {
		return summary, err
	}

	var mapped int
	r.getDB(ctx).Raw(`
		SELECT COUNT(DISTINCT al.employee_id) FROM attendance_logs al
		INNER JOIN employee_schedules es
			ON es.employee_id = al.employee_id
			AND es.is_active = TRUE
			AND es.effective_date <= ?::DATE
			AND (es.end_date IS NULL OR es.end_date >= ?::DATE)
			AND es.deleted_at IS NULL
		INNER JOIN shift_templates st ON st.id = es.shift_template_id AND st.deleted_at IS NULL
		INNER JOIN shift_template_details std
			ON std.shift_template_id = st.id
			AND std.day_of_week = LOWER(TRIM(TO_CHAR(?::DATE, 'Day')))::day_of_week_enum
			AND std.is_working_day = TRUE
			AND std.deleted_at IS NULL
		WHERE al.attendance_date = ?::DATE AND al.deleted_at IS NULL
	`, date, date, date, date).Scan(&mapped)

	summary.NotClockedIn = summary.TotalEmployees - mapped
	if summary.NotClockedIn < 0 {
		summary.NotClockedIn = 0
	}
	return summary, nil
}

func (r *dashboardRepository) GetTeamMutabaahSummary(ctx context.Context, date string) (dto.TeamMutabaahSummaryDTO, error) {
	var summary dto.TeamMutabaahSummaryDTO
	err := r.getDB(ctx).Raw(`
		SELECT
			(SELECT COUNT(DISTINCT employee_id) FROM attendance_logs
			 WHERE attendance_date = ?::DATE AND status IN ('present', 'late') AND deleted_at IS NULL) AS total_employees,
			(SELECT COUNT(DISTINCT employee_id) FROM mutabaah_logs
			 WHERE log_date = ?::DATE AND is_submitted = TRUE AND deleted_at IS NULL)                  AS submitted_count
	`, date, date).Scan(&summary).Error
	if err == nil {
		summary.NotSubmittedCount = summary.TotalEmployees - summary.SubmittedCount
		if summary.NotSubmittedCount < 0 {
			summary.NotSubmittedCount = 0
		}
	}
	return summary, err
}

func (r *dashboardRepository) GetNotClockedIn(ctx context.Context, date string) ([]dto.NotClockedInDTO, error) {
	var list []dto.NotClockedInDTO
	err := r.getDB(ctx).Raw(`
		SELECT
			e.id             AS employee_id,
			e.full_name      AS employee_name,
			e.employee_number,
			d.name           AS department_name,
			std.clock_in_start AS shift_start
		FROM employees e
		LEFT JOIN departments d ON d.id = e.department_id AND d.deleted_at IS NULL
		LEFT JOIN employee_schedules es
			ON es.employee_id = e.id
			AND es.is_active = TRUE
			AND es.effective_date <= ?::DATE
			AND (es.end_date IS NULL OR es.end_date >= ?::DATE)
			AND es.deleted_at IS NULL
		LEFT JOIN shift_templates st ON st.id = es.shift_template_id AND st.deleted_at IS NULL
		LEFT JOIN shift_template_details std
			ON std.shift_template_id = st.id
			AND std.day_of_week = LOWER(TRIM(TO_CHAR(?::DATE, 'Day')))::day_of_week_enum
			AND std.is_working_day = TRUE
			AND std.deleted_at IS NULL
		WHERE e.deleted_at IS NULL
		  AND std.clock_in_start IS NOT NULL
		  AND e.id NOT IN (
			SELECT employee_id FROM attendance_logs
			WHERE attendance_date = ?::DATE AND deleted_at IS NULL
		  )
		ORDER BY std.clock_in_start ASC
		LIMIT 10
	`, date, date, date, date).Scan(&list).Error
	return list, err
}

func (r *dashboardRepository) GetExpiringContracts(ctx context.Context, days int) ([]dto.ExpiringContractDTO, error) {
	var list []dto.ExpiringContractDTO
	if days <= 0 {
		return nil, errors.New("days must be positive")
	}
	query := fmt.Sprintf(`
		SELECT
			e.id             AS employee_id,
			e.full_name      AS employee_name,
			e.employee_number,
			ec.contract_type::TEXT AS contract_type,
			ec.end_date::TEXT      AS end_date,
			(ec.end_date - CURRENT_DATE) AS days_remaining
		FROM employment_contracts ec
		JOIN employees e ON e.id = ec.employee_id AND e.deleted_at IS NULL
		WHERE ec.end_date BETWEEN CURRENT_DATE AND (CURRENT_DATE + INTERVAL '%d days')
		  AND ec.start_date <= CURRENT_DATE
		  AND ec.deleted_at IS NULL
		ORDER BY ec.end_date ASC
	`, days)
	err := r.getDB(ctx).Raw(query).Scan(&list).Error
	return list, err
}

func (r *dashboardRepository) GetFastestArrivalRanking(ctx context.Context, date string, limit int) ([]dto.RankingEntryDTO, error) {
	var list []dto.RankingEntryDTO
	query := `
		WITH ActiveSchedule AS (
			SELECT DISTINCT ON (es.employee_id)
				es.employee_id,
				es.shift_template_id
			FROM employee_schedules es
			WHERE es.is_active = TRUE
			AND es.deleted_at IS NULL
			ORDER BY es.employee_id, es.effective_date DESC
		),
		RankedArrivals AS (
			SELECT
				al.employee_id,
				e.full_name,
				e.employee_number,
				al.clock_in_at,
				EXTRACT(EPOCH FROM (
					al.clock_in_at::time - COALESCE(std.clock_in_end, '08:00:00')::time
				)) / 60.0 AS diff_minutes
			FROM attendance_logs al
			JOIN employees e ON e.id = al.employee_id
			LEFT JOIN ActiveSchedule acs ON acs.employee_id = al.employee_id
			LEFT JOIN shift_template_details std
				ON std.shift_template_id = acs.shift_template_id
				AND std.day_of_week = LOWER(TRIM(TO_CHAR(al.attendance_date, 'Day')))::day_of_week_enum
				AND std.is_working_day = TRUE
				AND std.deleted_at IS NULL
			WHERE al.attendance_date = ?::DATE
			AND al.status IN ('present', 'late')
			AND al.deleted_at IS NULL
		)
		SELECT
			RANK() OVER (ORDER BY diff_minutes ASC) AS rank,
			employee_id,
			full_name AS employee_name,
			employee_number,
			ROUND(diff_minutes::numeric, 0) AS value,
			TO_CHAR(clock_in_at, 'HH24:MI') AS value_label
		FROM RankedArrivals
		WHERE diff_minutes < 0
		ORDER BY diff_minutes ASC
		LIMIT ?
	`
	err := r.getDB(ctx).Raw(query, date, limit).Scan(&list).Error
	return list, err
}

func (r *dashboardRepository) GetTopTilawahByDepartment(ctx context.Context, date string, limit int) ([]dto.DepartmentRankingDTO, error) {
	var list []dto.DepartmentRankingDTO
	query := `
		WITH EligibleEmployees AS (
			SELECT DISTINCT e.id AS employee_id, e.department_id
			FROM employees e
			JOIN accounts a ON a.employee_id = e.id AND a.deleted_at IS NULL
			JOIN role_permissions rp ON rp.role_id = a.role_id AND rp.permission_code = 'mutabaah-create'
			WHERE e.deleted_at IS NULL AND e.department_id IS NOT NULL
		),
		DeptTilawah AS (
			SELECT
				d.id AS department_id,
				d.name AS department_name,
				COUNT(ee.employee_id) AS total_eligible,
				COUNT(ml.id) FILTER (WHERE ml.is_submitted = TRUE) AS submitted_count
			FROM EligibleEmployees ee
			JOIN departments d ON d.id = ee.department_id AND d.deleted_at IS NULL
			LEFT JOIN mutabaah_logs ml ON ml.employee_id = ee.employee_id
				AND ml.log_date = ?::DATE
				AND ml.deleted_at IS NULL
			GROUP BY d.id, d.name
			HAVING COUNT(ee.employee_id) > 0
		)
		SELECT
			RANK() OVER (ORDER BY (submitted_count::float / total_eligible) DESC) AS rank,
			department_id,
			department_name,
			ROUND((submitted_count::numeric / total_eligible) * 100, 1) AS value,
			ROUND((submitted_count::numeric / total_eligible) * 100, 1) || '%' AS value_label
		FROM DeptTilawah
		ORDER BY value DESC
		LIMIT ?
	`
	err := r.getDB(ctx).Raw(query, date, limit).Scan(&list).Error
	return list, err
}

func (r *dashboardRepository) GetFastestMutabaahRanking(ctx context.Context, date string, limit int) ([]dto.RankingEntryDTO, error) {
	var list []dto.RankingEntryDTO
	query := `
		WITH RankedMutabaah AS (
			SELECT
				ml.employee_id,
				e.full_name AS employee_name,
				e.employee_number,
				EXTRACT(EPOCH FROM (ml.submitted_at - al.clock_in_at)) / 60.0 AS value,
				TO_CHAR(ml.submitted_at, 'HH24:MI') AS value_label
			FROM mutabaah_logs ml
			JOIN employees e ON e.id = ml.employee_id
			JOIN attendance_logs al ON al.id = ml.attendance_log_id AND al.deleted_at IS NULL
			WHERE ml.log_date = ?::DATE
			  AND ml.is_submitted = TRUE
			  AND ml.deleted_at IS NULL
		)
		SELECT
			RANK() OVER (ORDER BY value ASC) AS rank,
			employee_id,
			employee_name,
			employee_number,
			value,
			value_label
		FROM RankedMutabaah
		ORDER BY value ASC
		LIMIT ?
	`
	err := r.getDB(ctx).Raw(query, date, limit).Scan(&list).Error
	return list, err
}

func (r *dashboardRepository) GetRecentAttendanceMeta(ctx context.Context, employeeID *uint) ([]dto.Meta, error) {
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

	err := r.db.Raw(query, args...).Scan(&meta).Error
	return meta, err
}

func (r *dashboardRepository) GetLeaveTypeMeta(ctx context.Context) ([]dto.Meta, error) {
	var meta []dto.Meta
	err := r.getDB(ctx).Raw("SELECT id::TEXT AS id, name FROM leave_types WHERE deleted_at IS NULL ORDER BY name ASC").Scan(&meta).Error
	return meta, err
}

func (r *dashboardRepository) GetEmployeeMeta(ctx context.Context, employeeID *uint) ([]dto.Meta, error) {
	var meta []dto.Meta
	var args []interface{}

	query := `
		SELECT 
			e.id::TEXT AS id,
			e.full_name AS name
		FROM employees e
		WHERE e.deleted_at IS NULL
	`
	if employeeID != nil {
		query += " AND e.id = ?"
		args = []interface{}{*employeeID}
	} else {
		query += " ORDER BY e.full_name ASC"
	}
	err := r.db.Raw(query, args...).Scan(&meta).Error
	return meta, err
}