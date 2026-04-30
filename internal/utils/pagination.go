package utils

import (
	"fmt"
	"hris-backend/internal/struct/dto"
)

// AllowedSortColumns — whitelist kolom yang boleh di-sort per domain
// Dipanggil dari masing-masing repository untuk validasi sort_by
var AllowedSortColumns = map[string]map[string]string{
	"attendance": {
		"attendance_date": "al.attendance_date",
		"employee_name":   "e.full_name",
		"status":          "al.status",
		"clock_in_at":     "al.clock_in_at",
		"clock_out_at":    "al.clock_out_at",
		"late_minutes":    "al.late_minutes",
		"department":      "d.name",
	},
	"employee": {
		"full_name":       "e.full_name",
		"employee_number": "e.employee_number",
		"department":      "d.name",
		"branch":          "b.name",
		"job_positions_id":"jp.title",
		"created_at":      "e.created_at",
	},
	"branches": {
		"name":       "name",
		"code":       "code",
		"created_at": "created_at",
	},
	"business_trip": {
		"employee_name": "e.full_name",
		"start_date":    "b.start_date",
		"end_date":      "b.end_date",
		"status":        "b.status",
		"destination":   "b.destination",
		"department":    "d.name",
		"created_at":    "b.created_at",
	},
	"daily_reports": {
		"report_date":    "d.report_date",
		"employee_name":  "e.full_name",
		"is_submitted":   "d.is_submitted",
		"created_at":     "d.created_at",
	},
	"departments": {
		"name":       "d.name",
		"created_at": "d.created_at",
	},
	"holiday": {
		"date":       "h.date",
		"name":       "h.name",
		"type":       "h.type",
		"created_at": "h.created_at",
	},
	"leave_balances": {
		"employee_name": "e.full_name",
		"leave_type":    "lt.name",
		"year":          "b.year",
		"used_duration": "b.used_duration",
		"created_at":    "b.created_at",
	},
	"leave_requests": {
		"employee_name": "e.full_name",
		"start_date":    "r.start_date",
		"end_date":      "r.end_date",
		"status":        "r.status",
		"leave_type":    "lt.name",
		"department":    "d.name",
		"total_days":    "r.total_days",
		"created_at":    "r.created_at",
	},
	"mutabaah": {
		"log_date":       "ml.log_date",
		"employee_name":  "e.full_name",
		"is_submitted":   "ml.is_submitted",
		"target_pages":   "ml.target_pages",
		"created_at":     "ml.created_at",
	},
	"overtime": {
		"employee_name": "e.full_name",
		"overtime_date": "o.overtime_date",
		"status":        "o.status",
		"duration":      "o.duration_minutes",
		"department":    "d.name",
		"created_at":    "o.created_at",
	},
	"permission": {
		"employee_name":   "e.full_name",
		"permission_date": "pr.permission_date",
		"permission_type": "pr.permission_type",
		"status":          "pr.status",
		"department":      "d.name",
		"created_at":      "pr.created_at",
	},
	"roles": {
		"name":       "r.name",
		"level":      "r.level",
		"created_at": "r.created_at",
	},
	"shift_templates": {
		"name":       "name",
		"created_at": "created_at",
	},
	"employee_schedules": {
		"employee_name":  "e.full_name",
		"shift_name":     "st.name",
		"effective_date": "es.effective_date",
		"is_active":      "es.is_active",
		"created_at":     "es.created_at",
	},
}

// BuildSortClause — generate ORDER BY clause yang aman dari SQL injection
// defaultSort sudah mengandung arah (e.g. "al.attendance_date DESC")
func BuildSortClause(domain string, sortBy *string, sortDir string, defaultSort string) string {
	if sortBy == nil || *sortBy == "" {
		return fmt.Sprintf(" ORDER BY %s", defaultSort)
	}
	cols, ok := AllowedSortColumns[domain]
	if !ok {
		return fmt.Sprintf(" ORDER BY %s", defaultSort)
	}
	col, ok := cols[*sortBy]
	if !ok {
		return fmt.Sprintf(" ORDER BY %s", defaultSort)
	}
	return fmt.Sprintf(" ORDER BY %s %s", col, sortDir)
}

// BuildPaginationClause — generate LIMIT/OFFSET clause
func BuildPaginationClause(p dto.PaginationParams) string {
	perPage := p.GetPerPage()
	if perPage == 0 { // semua
		return ""
	}
	page := p.GetPage()
	offset := (page - 1) * perPage
	return fmt.Sprintf(" LIMIT %d OFFSET %d", perPage, offset)
}
