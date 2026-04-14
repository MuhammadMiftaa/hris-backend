package repository

import (
	"context"

	"gorm.io/gorm"
)

type Transaction interface {
	Commit() error
	Rollback() error
}

type GormTx struct {
	db *gorm.DB
}

func (txm *GormTx) Commit() error {
	return txm.db.Commit().Error
}

func (txm *GormTx) Rollback() error {
	return txm.db.Rollback().Error
}

type TxManager interface {
	Begin(ctx context.Context) (Transaction, error)
}


type GormTxManager struct {
	db *gorm.DB
}

func (txm *GormTxManager) Begin(ctx context.Context) (Transaction, error) {
	tx := txm.db.Begin().WithContext(ctx)
	return &GormTx{db: tx}, tx.Error
}

func NewTxManager(db *gorm.DB) TxManager {
	return &GormTxManager{db: db}
}