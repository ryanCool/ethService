package postgres

import (
	"context"
	"github.com/ryanCool/ethService/domain"
	"gorm.io/gorm"
)

type postgresBlockRepository struct {
	Db *gorm.DB
}

// NewPostgresBlockRepository will create an object that represent the block.Repository interface
func NewPostgresBlockRepository(db *gorm.DB) domain.BlockRepository {
	return &postgresBlockRepository{db}
}

func (p *postgresBlockRepository) Create(ctx context.Context, block *domain.BlockDb) error {
	return p.Db.Table("eth.blocks").Create(&block).Error
}

func (p *postgresBlockRepository) List(ctx context.Context, limit int) ([]domain.BlockDb, error) {
	var res []domain.BlockDb
	if err := p.Db.Table("eth.blocks").Order("block_num desc").Find(&res).Error; err != nil {
		return nil, err
	}
	return res, nil
}
