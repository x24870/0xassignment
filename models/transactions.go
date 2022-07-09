package models

import (
	"github.com/jinzhu/gorm"
)

// TransactionIntf ...
type TransactionIntf interface {
	GetID() uint64
	GetBlockHash() string
	GetTxHash() string
	GetTxFrom() string
	GetTxTo() string
	GetNounce() uint64
	GetData() []byte
	GetValue() string
	GetCreatedAt() int64
	GetUpdatedAt() int64
	GetByHash(db *gorm.DB, hash string) (TransactionIntf, error)
}

// Transaction is the exported static model interface.
var Transaction transaction

// transaction ...
type transaction struct {
	ID        int    `gorm:"column:id;primary_key"`
	BlockHash string `gorm:"column:block_hash"`
	TxHash    string `gorm:"column:tx_hash"`
	TxFrom    string `gorm:"column:tx_from"`
	TxTo      string `gorm:"column:tx_to"`
	Nounce    uint64 `gorm:"column:nounce"`
	Data      []byte `gorm:"column:data"`
	Value     string `gorm:"column:value"`
	CreatedAt int64  `gorm:"column:created_at;default:extract(epoch from now())*1000"`
	UpdatedAt int64  `gorm:"column:updated_at;default:extract(epoch from now())*1000"`
}

func init() {
	registerModelForAutoMigration(&transaction{})
}

// TableName is used by GORM to choose which table to use.
func (b *transaction) TableName() string {
	return "transactions"
}

// createIndexes ...
func (b *transaction) createIndexes(db *gorm.DB) error {
	return nil
}

// createUniqueIndexes ...
func (b *transaction) createUniqueIndexes(db *gorm.DB) error {
	return nil
}

// createForeignKeys ...
func (b *transaction) createForeignKeys(db *gorm.DB) error {
	return nil
}

// GetID ...
func (b *transaction) GetID() uint64 {
	return uint64(b.ID)
}

// GetBlockHash ...
func (b *transaction) GetBlockHash() string {
	return b.BlockHash
}

// GetTxHash ...
func (b *transaction) GetTxHash() string {
	return b.TxHash
}

// GetTxFrom ...
func (b *transaction) GetTxFrom() string {
	return b.TxFrom
}

// GetTxTo ...
func (b *transaction) GetTxTo() string {
	return b.TxTo
}

// GetNounce ...
func (b *transaction) GetNounce() uint64 {
	return b.Nounce
}

// GetData ...
func (b *transaction) GetData() []byte {
	return b.Data
}

// GetValue ...
func (b *transaction) GetValue() string {
	return b.Value
}

// GetCreatedAt ...
func (b *transaction) GetCreatedAt() int64 {
	return b.CreatedAt
}

// GetUpdatedAt ...
func (b *transaction) GetUpdatedAt() int64 {
	return b.UpdatedAt
}

// GetByHash ...
func (b *transaction) GetByHash(db *gorm.DB, hash string) (TransactionIntf, error) {
	// Get transaction based on given number
	transaction := transaction{}
	err := db.Model(b).Where("number = ?", hash).First(&transaction).Error
	if err != nil {
		return nil, err
	}

	return &transaction, nil
}
