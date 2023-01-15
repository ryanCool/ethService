package postgres

import (
	"context"
	"github.com/ryanCool/ethService/domain"
	"gorm.io/gorm"
)

type postgresTransactionRepository struct {
	Db *gorm.DB
}

// NewPostgresTransactionRepository will create an object that represent the transaction.Repository interface
func NewPostgresTransactionRepository(db *gorm.DB) domain.TransactionRepository {
	return &postgresTransactionRepository{db}
}

func (p *postgresTransactionRepository) Create(ctx context.Context, transaction *domain.Transaction) error {
	return p.Db.Table("eth.transactions").Create(&transaction).Error
}
