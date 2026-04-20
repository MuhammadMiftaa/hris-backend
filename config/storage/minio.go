package storage

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"hris-backend/config/env"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/cors"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const (
	BucketAttendance    = "attendance-photos"
	BucketMutabaah      = "mutabaah-docs"
	BucketProfilePhotos = "profile-photos"

	// Presigned PUT URL untuk upload foto dari browser (kamera PWA) — 5 menit
	PresignedUploadExpiry = 5 * time.Minute

	// Presigned GET URL untuk akses foto — 15 menit, tidak bisa disimpan/dibagikan permanen
	PresignedDownloadExpiry = 15 * time.Minute
)

// MinioClient — interface untuk abstraksi dan mocking
type MinioClient interface {
	PresignedPutObject(ctx context.Context, bucket, object string, expiry time.Duration) (string, error)
	PresignedGetObject(ctx context.Context, bucket, object string, expiry time.Duration) (string, error)
	EnsureBuckets(ctx context.Context) error
}

type minioClient struct {
	client         *minio.Client
	internalOrigin string
	publicOrigin   string
}

// NewMinioClient membuat koneksi ke MinIO server menggunakan config env
func NewMinioClient(cfg env.Minio) (MinioClient, error) {
	endpoint := cfg.Host
	if cfg.Port != "" {
		endpoint = fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.Secure,
	})
	if err != nil {
		return nil, fmt.Errorf("minio: failed to create client: %w", err)
	}

	scheme := "http"
	if cfg.Secure {
		scheme = "https"
	}
	internalOrigin := fmt.Sprintf("%s://%s", scheme, endpoint)

	return &minioClient{
		client:         client,
		internalOrigin: internalOrigin,
		publicOrigin:   cfg.PublicURL,
	}, nil
}

// EnsureBuckets buat bucket saat startup jika belum ada, dan set policy private
func (m *minioClient) EnsureBuckets(ctx context.Context) error {
	buckets := []string{BucketAttendance, BucketMutabaah, BucketProfilePhotos}
	for _, bucket := range buckets {
		exists, err := m.client.BucketExists(ctx, bucket)
		if err != nil {
			return fmt.Errorf("minio: check bucket %s: %w", bucket, err)
		}
		if !exists {
			if err := m.client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
				return fmt.Errorf("minio: create bucket %s: %w", bucket, err)
			}
			policy := fmt.Sprintf(`{...}`, bucket)
			if err := m.client.SetBucketPolicy(ctx, bucket, policy); err != nil {
				fmt.Printf("minio: warn: set policy %s: %v\n", bucket, err)
			}
		}

		corsConfig := cors.NewConfig([]cors.Rule{
			{
				AllowedOrigin: []string{"https://hris.miv.best"},
				AllowedMethod: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
				AllowedHeader: []string{"*"},
				ExposeHeader:  []string{"ETag"},
				MaxAgeSeconds: 3600,
			},
		})
		if err := m.client.SetBucketCors(ctx, bucket, corsConfig); err != nil {
			fmt.Printf("minio: warn: set cors %s: %v\n", bucket, err)
		}
	}
	return nil
}

func (m *minioClient) toPublicURL(u *url.URL) string {
	if m.publicOrigin == "" {
		return u.String()
	}
	return strings.Replace(u.String(), m.internalOrigin, m.publicOrigin, 1)
}

func (m *minioClient) PresignedPutObject(ctx context.Context, bucket, object string, expiry time.Duration) (string, error) {
	u, err := m.client.PresignedPutObject(ctx, bucket, object, expiry)
	if err != nil {
		return "", fmt.Errorf("minio: presigned put %s/%s: %w", bucket, object, err)
	}
	return m.toPublicURL(u), nil
}

func (m *minioClient) PresignedGetObject(ctx context.Context, bucket, object string, expiry time.Duration) (string, error) {
	u, err := m.client.PresignedGetObject(ctx, bucket, object, expiry, url.Values{})
	if err != nil {
		return "", fmt.Errorf("minio: presigned get %s/%s: %w", bucket, object, err)
	}
	return m.toPublicURL(u), nil
}
