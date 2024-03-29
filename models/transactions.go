package models

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/jinzhu/gorm"
)

// TransactionIntf ...
type TransactionIntf interface {
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
	GetByBlockHash(db *gorm.DB, hash string) ([]TransactionIntf, error)
	SetTransaction(db *gorm.DB) error
}

// Transaction is the exported static model interface.
var Transaction transaction

// transaction ...
type transaction struct {
	BlockHash string `gorm:"column:block_hash" json:"-"`
	TxHash    string `gorm:"column:tx_hash" json:"tx_hash"`
	TxFrom    string `gorm:"column:tx_from" json:"from"`
	TxTo      string `gorm:"column:tx_to" json:"to"`
	Nounce    uint64 `gorm:"column:nounce" json:"nounce"`
	Data      []byte `gorm:"column:data" json:"data"`
	Value     string `gorm:"column:value" json:"value"`
	CreatedAt int64  `gorm:"column:created_at;default:extract(epoch from now())*1000" json:"-"`
	UpdatedAt int64  `gorm:"column:updated_at;default:extract(epoch from now())*1000" json:"-"`
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

// NewTransaction
func NewTransaction(t *types.Transaction, blockHash string) (TransactionIntf, error) {
	from, err := types.Sender(types.LatestSignerForChainID(t.ChainId()), t)
	if err != nil {
		return nil, err
	}
	to := t.To()
	if to == nil {
		to = &common.Address{}
	}
	newTransaction := transaction{
		BlockHash: blockHash,
		TxHash:    t.Hash().String(),
		TxFrom:    from.String(),
		TxTo:      to.String(),
		Nounce:    t.Nonce(),
		Data:      t.Data(),
		Value:     t.Value().String(),
	}

	return &newTransaction, nil
}

// GetByHash ...
func (t *transaction) GetByHash(db *gorm.DB, hash string) (TransactionIntf, error) {
	// Get transaction based on given number
	transaction := transaction{}
	err := db.Model(t).Where("tx_hash = ?", hash).First(&transaction).Error
	if err != nil {
		return nil, err
	}

	return &transaction, nil
}

// GetByBlockHash ...
func (t *transaction) GetByBlockHash(db *gorm.DB, hash string) ([]TransactionIntf, error) {
	transactions := []*transaction{}
	err := db.Model(t).Where("block_hash = ?", hash).Find(&transactions).Error
	if err != nil {
		return nil, err
	}

	// Organize into ServiceIntf slice.
	transactionIntfs := []TransactionIntf{}
	for _, t := range transactions {
		transactionIntfs = append(transactionIntfs, t)
	}

	return transactionIntfs, nil
}

// SetTransaction ...
func (t *transaction) SetTransaction(db *gorm.DB) error {
	return db.Where("tx_hash = ?", t.TxHash).FirstOrCreate(t).Error
}
