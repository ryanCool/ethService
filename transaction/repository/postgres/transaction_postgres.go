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

func (p *postgresTransactionRepository) SaveReceiptAndLogs(ctx context.Context, txHash string, logs []domain.TransactionLog) error {
	return p.Db.Transaction(func(tx *gorm.DB) error {

		if err := tx.Table("eth.receipts").Create(&domain.Receipt{TxHash: txHash}).Error; err != nil {
			return err
		}

		if err := tx.Table("eth.transaction_logs").Create(&logs).Error; err != nil {
			return err
		}

		return nil
	})
}

func (p *postgresTransactionRepository) GetLogsByTxHash(ctx context.Context, txHash string) ([]domain.TransactionLog, error) {
	var res []domain.TransactionLog
	if err := p.Db.Table("eth.transaction_logs").Where("tx_hash = ?", txHash).Find(&res).Error; err != nil {
		return nil, err
	}

	return res, nil
}

func (p *postgresTransactionRepository) GetByTxHash(ctx context.Context, txHash string) (*domain.Transaction, error) {
	var res *domain.Transaction
	if err := p.Db.Table("eth.transactions").Where("tx_hash = ?", txHash).First(&res).Error; err != nil {
		return nil, err
	}

	return res, nil
}
