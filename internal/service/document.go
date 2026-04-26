package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"hris-backend/config/storage"
	"hris-backend/internal/struct/dto"
	"hris-backend/internal/utils"
)

type DocumentService interface {
	UploadDocument(ctx context.Context, employeeID uint, req dto.UploadDocumentRequest) (dto.UploadDocumentResponse, error)
	GetDownloadURL(ctx context.Context, req dto.DocumentDownloadRequest) (dto.DocumentDownloadResponse, error)
}

type documentService struct {
	minio storage.MinioClient
}

func NewDocumentService(minio storage.MinioClient) DocumentService {
	return &documentService{minio: minio}
}

func (s *documentService) UploadDocument(ctx context.Context, employeeID uint, req dto.UploadDocumentRequest) (dto.UploadDocumentResponse, error) {
	// Decode base64
	base64Data := req.Base64Document
	if idx := strings.Index(base64Data, ","); idx != -1 {
		base64Data = base64Data[idx+1:]
	}

	docBytes, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return dto.UploadDocumentResponse{}, fmt.Errorf("gagal decode base64 document: %w", err)
	}

	// Max 3MB
	if len(docBytes) > 3*1024*1024 {
		return dto.UploadDocumentResponse{}, fmt.Errorf("ukuran dokumen maksimal 3MB")
	}

	// Deteksi extensi dari filename
	ext := "bin"
	if idx := strings.LastIndex(req.Filename, "."); idx != -1 {
		ext = req.Filename[idx+1:]
	}
	ext = strings.ToLower(ext)

	allowedExts := map[string]bool{
		"pdf": true, "doc": true, "docx": true, "jpg": true, "jpeg": true, "png": true,
	}
	if !allowedExts[ext] {
		return dto.UploadDocumentResponse{}, fmt.Errorf("ekstensi dokumen tidak diizinkan")
	}

	var bucket string
	if req.DocumentType == "leave" {
		bucket = storage.BucketLeaveDocuments
	} else if req.DocumentType == "business_trip" {
		bucket = storage.BucketBusinessTripDocuments
	} else {
		return dto.UploadDocumentResponse{}, fmt.Errorf("tipe dokumen tidak valid")
	}

	// Upload to MinIO
	objectKey := fmt.Sprintf("%d/%d_%s", employeeID, time.Now().Unix(), req.Filename)

	uploadURL, err := s.minio.PresignedPutObject(ctx, bucket, objectKey, storage.PresignedUploadExpiry)
	if err != nil {
		return dto.UploadDocumentResponse{}, fmt.Errorf("gagal membuat upload URL: %w", err)
	}

	contentType := "application/octet-stream"
	switch ext {
	case "pdf":
		contentType = "application/pdf"
	case "png":
		contentType = "image/png"
	case "jpg", "jpeg":
		contentType = "image/jpeg"
	case "doc":
		contentType = "application/msword"
	case "docx":
		contentType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	}

	if err := utils.UploadToPresignedURL(uploadURL, docBytes, contentType); err != nil {
		return dto.UploadDocumentResponse{}, fmt.Errorf("gagal upload dokumen: %w", err)
	}

	return dto.UploadDocumentResponse{
		Success:     true,
		DocumentURL: objectKey,
		Message:     "Dokumen berhasil diupload",
	}, nil
}

func (s *documentService) GetDownloadURL(ctx context.Context, req dto.DocumentDownloadRequest) (dto.DocumentDownloadResponse, error) {
	var bucket string
	if req.DocumentType == "leave" {
		bucket = storage.BucketLeaveDocuments
	} else if req.DocumentType == "business_trip" {
		bucket = storage.BucketBusinessTripDocuments
	} else {
		return dto.DocumentDownloadResponse{}, fmt.Errorf("tipe dokumen tidak valid")
	}

	url, err := s.minio.PresignedGetObject(ctx, bucket, req.Key, storage.PresignedDownloadExpiry)
	if err != nil {
		return dto.DocumentDownloadResponse{}, fmt.Errorf("gagal generate download URL: %w", err)
	}

	return dto.DocumentDownloadResponse{
		URL: url,
	}, nil
}
