package dto

type ExportFormat string

const (
	ExportCSV ExportFormat = "csv"
	ExportPDF ExportFormat = "pdf"
)

type ExportRequest struct {
	Format ExportFormat `query:"format"` // "csv" atau "pdf"
}
