package dto

// PaginationParams — query parameter pagination & sorting universal
type PaginationParams struct {
	Page     *int    `query:"page"`      // default: 1
	PerPage  *int    `query:"per_page"`  // default: 10, opsi: 10,25,50,100,0(semua)
	SortBy   *string `query:"sort_by"`   // nama kolom, e.g. "attendance_date"
	SortDir  *string `query:"sort_dir"`  // "asc" atau "desc", default: "desc"
}

// PaginatedResponse — wrapper response dengan metadata pagination
type PaginatedResponse[T any] struct {
	Data       []T            `json:"data"`
	Pagination PaginationMeta `json:"pagination"`
}

type PaginationMeta struct {
	Page      int `json:"page"`
	PerPage   int `json:"per_page"`
	Total     int `json:"total"`
	TotalPage int `json:"total_page"`
}

// Defaults — return safe defaults
func (p PaginationParams) GetPage() int {
	if p.Page == nil || *p.Page < 1 {
		return 1
	}
	return *p.Page
}

func (p PaginationParams) GetPerPage() int {
	if p.PerPage == nil {
		return 10
	}
	if *p.PerPage == 0 { // 0 = semua
		return 0
	}
	allowed := map[int]bool{10: true, 25: true, 50: true, 100: true}
	if allowed[*p.PerPage] {
		return *p.PerPage
	}
	return 10
}

func (p PaginationParams) GetSortDir() string {
	if p.SortDir != nil && *p.SortDir == "asc" {
		return "ASC"
	}
	return "DESC"
}
