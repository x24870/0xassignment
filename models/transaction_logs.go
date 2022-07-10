package models

import (
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/jinzhu/gorm"
)

// TransactionLogIntf ...
type TransactionLogIntf interface {
	GetTxHash() string
	GetByHash(db *gorm.DB, txHash string) ([]TransactionLogIntf, error)
	SetTransactionLog(db *gorm.DB) error
}

// TransactionLog is the exported static model interface.
var TransactionLog transactionLog

// transactionLog ...
type transactionLog struct {
	TxHash    string `gorm:"column:tx_hash" json:"tx_hash"`
	LogIndex  int64  `gorm:"column:log_index" json:"log_index"`
	Data      []byte `gorm:"column:data" json:"data"`
	CreatedAt int64  `gorm:"column:created_at;default:extract(epoch from now())*1000" json:"-"`
	UpdatedAt int64  `gorm:"column:updated_at;default:extract(epoch from now())*1000" json:"-"`
}

func init() {
	registerModelForAutoMigration(&transactionLog{})
}

// TableName is used by GORM to choose which table to use.
func (t *transactionLog) TableName() string {
	return "transaction_logs"
}

// createIndexes ...
func (t *transactionLog) createIndexes(db *gorm.DB) error {
	return nil
}

// createUniqueIndexes ...
func (t *transactionLog) createUniqueIndexes(db *gorm.DB) error {
	return nil
}

// createForeignKeys ...
func (t *transactionLog) createForeignKeys(db *gorm.DB) error {
	return nil
}

// GetTxHash ...
func (t *transactionLog) GetTxHash() string {
	return t.TxHash
}

// GetCreatedAt ...
func (t *transactionLog) GetCreatedAt() int64 {
	return t.CreatedAt
}

// GetUpdatedAt ...
func (t *transactionLog) GetUpdatedAt() int64 {
	return t.UpdatedAt
}

// NewTransactionLog
func NewTransactionLog(t *types.Log, txHash string) (TransactionLogIntf, error) {
	newTransactionLog := transactionLog{
		TxHash: txHash,
	}

	return &newTransactionLog, nil
}

// GetByHash ...
func (t *transactionLog) GetByHash(db *gorm.DB, txHash string) ([]TransactionLogIntf, error) {
	// Get transactionLog based on given transaction hash
	transactionLogs := []*transactionLog{}
	err := db.Model(t).Where("tx_hash = ?", txHash).First(&transactionLogs).Error
	if err != nil {
		return nil, err
	}

	// Organize into ServiceIntf slice.
	transactionLogIntfs := []TransactionLogIntf{}
	for _, t := range transactionLogs {
		transactionLogIntfs = append(transactionLogIntfs, t)
	}

	return transactionLogIntfs, nil
}

// SetTransactionLog ...
func (t *transactionLog) SetTransactionLog(db *gorm.DB) error {
	return db.Where("tx_hash = ?", t.TxHash).FirstOrCreate(t).Error
}
