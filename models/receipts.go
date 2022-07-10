package models

import (
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/jinzhu/gorm"
)

// ReceiptIntf ...
type ReceiptIntf interface {
	GetTxHash() string
	GetByHash(db *gorm.DB, txHash string) (ReceiptIntf, error)
	SetReceipt(db *gorm.DB) error
}

// Receipt is the exported static model interface.
var Receipt receipt

// receipt ...
type receipt struct {
	TxHash    string `gorm:"column:tx_hash" json:"tx_hash"`
	CreatedAt int64  `gorm:"column:created_at;default:extract(epoch from now())*1000" json:"-"`
	UpdatedAt int64  `gorm:"column:updated_at;default:extract(epoch from now())*1000" json:"-"`
}

func init() {
	registerModelForAutoMigration(&receipt{})
}

// TableName is used by GORM to choose which table to use.
func (r *receipt) TableName() string {
	return "receipts"
}

// createIndexes ...
func (r *receipt) createIndexes(db *gorm.DB) error {
	return nil
}

// createUniqueIndexes ...
func (r *receipt) createUniqueIndexes(db *gorm.DB) error {
	return nil
}

// createForeignKeys ...
func (r *receipt) createForeignKeys(db *gorm.DB) error {
	return nil
}

// GetTxHash ...
func (r *receipt) GetTxHash() string {
	return r.TxHash
}

// GetCreatedAt ...
func (r *receipt) GetCreatedAt() int64 {
	return r.CreatedAt
}

// GetUpdatedAt ...
func (r *receipt) GetUpdatedAt() int64 {
	return r.UpdatedAt
}

// NewReceipt
func NewReceipt(r *types.Receipt, txHash string) (ReceiptIntf, error) {
	newReceipt := receipt{
		TxHash: txHash,
	}

	return &newReceipt, nil
}

// GetByHash ...
func (r *receipt) GetByHash(db *gorm.DB, txHash string) (ReceiptIntf, error) {
	// Get receipt based on given transaction hash
	receipt := receipt{}
	err := db.Model(r).Where("tx_hash = ?", txHash).First(&receipt).Error
	if err != nil {
		return nil, err
	}

	return &receipt, nil
}

// SetReceipt ...
func (r *receipt) SetReceipt(db *gorm.DB) error {
	return db.Where("tx_hash = ?", r.TxHash).FirstOrCreate(r).Error
}
