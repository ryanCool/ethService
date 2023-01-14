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

func (p *postgresBlockRepository) Create(ctx context.Context, block *domain.Block) error {
	return p.Db.Table("eth.blocks").Create(&block).Error
}
