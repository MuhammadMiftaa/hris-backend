package repository

import (
	"context"
	"errors"

	"hris-backend/internal/struct/model"

	"gorm.io/gorm"
)

type PushRepository interface {
	SaveSubscription(ctx context.Context, tx Transaction, employeeID uint, endpoint, p256dh, auth, userAgent string) error
	GetActiveSubscriptionsByEmployeeID(ctx context.Context, tx Transaction, employeeID uint) ([]model.PushSubscription, error)
	DeactivateSubscription(ctx context.Context, tx Transaction, endpoint string) error
	IsEmployeeSubscribed(ctx context.Context, tx Transaction, employeeID uint) (bool, error)
}

type pushRepository struct {
	db *gorm.DB
}

func NewPushRepository(db *gorm.DB) PushRepository {
	return &pushRepository{db: db}
}

func (r *pushRepository) getDB(ctx context.Context, tx Transaction) (*gorm.DB, error) {
	if tx != nil {
		gormTx, ok := tx.(*GormTx)
		if !ok {
			return nil, errors.New("invalid transaction type")
		}
		return gormTx.db.WithContext(ctx), nil
	}
	return r.db.WithContext(ctx), nil
}

func (r *pushRepository) SaveSubscription(ctx context.Context, tx Transaction, employeeID uint, endpoint, p256dh, auth, userAgent string) error {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return err
	}

	// Soft-delete subscription lama dengan endpoint yang sama, lalu buat baru
	_ = db.Where("endpoint = ?", endpoint).Delete(&model.PushSubscription{})

	sub := model.PushSubscription{
		EmployeeID: employeeID,
		Endpoint:   endpoint,
		P256dh:     p256dh,
		Auth:       auth,
		IsActive:   true,
	}
	if userAgent != "" {
		sub.UserAgent = &userAgent
	}
	return db.Create(&sub).Error
}

func (r *pushRepository) GetActiveSubscriptionsByEmployeeID(ctx context.Context, tx Transaction, employeeID uint) ([]model.PushSubscription, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}

	var subs []model.PushSubscription
	err = db.Where("employee_id = ? AND is_active = TRUE AND deleted_at IS NULL", employeeID).Find(&subs).Error
	return subs, err
}

func (r *pushRepository) DeactivateSubscription(ctx context.Context, tx Transaction, endpoint string) error {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return err
	}
	return db.Model(&model.PushSubscription{}).
		Where("endpoint = ?", endpoint).
		Update("is_active", false).Error
}

func (r *pushRepository) IsEmployeeSubscribed(ctx context.Context, tx Transaction, employeeID uint) (bool, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return false, err
	}

	var count int64
	err = db.Model(&model.PushSubscription{}).
		Where("employee_id = ? AND is_active = TRUE AND deleted_at IS NULL", employeeID).
		Count(&count).Error
	return count > 0, err
}
