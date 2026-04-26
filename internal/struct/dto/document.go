package dto

type UploadDocumentRequest struct {
	Base64Document string `json:"base64_document" validate:"required"`
	Filename       string `json:"filename" validate:"required"`
	DocumentType   string `json:"document_type" validate:"required,oneof=leave business_trip"`
}

type UploadDocumentResponse struct {
	Success     bool   `json:"success"`
	DocumentURL string `json:"document_url"` // object key
	Message     string `json:"message"`
}

type DocumentDownloadRequest struct {
	Key          string `query:"key" validate:"required"`
	DocumentType string `query:"document_type" validate:"required,oneof=leave business_trip"`
}

type DocumentDownloadResponse struct {
	URL string `json:"url"` // presigned download URL
}
