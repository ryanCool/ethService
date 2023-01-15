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

func (p *postgresTransactionRepository) GetTxHashesByBlockHash(ctx context.Context, blockHash string) ([]string, error) {
	var res []domain.Transaction
	if err := p.Db.Select("tx_hash").Table("eth.transactions").Where("block_hash = ?", blockHash).Find(&res).Error; err != nil {
		return nil, err
	}

	var hashes []string
	for _, transaction := range res {
		hashes = append(hashes, transaction.TxHash)
	}

	return hashes, nil
}
