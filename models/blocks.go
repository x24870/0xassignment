package models

import (
	"errors"

	"github.com/ethereum/go-ethereum/core/types"

	"github.com/jinzhu/gorm"
)

// BlockIntf ...
type BlockIntf interface {
	GetNumber() uint64
	GetHash() string
	GetTime() uint64
	GetParent() string
	GetStable() bool
	GetCreatedAt() int64
	GetUpdatedAt() int64
	GetBlocks(db *gorm.DB, num uint64) ([]BlockIntf, error)
	GetByNumber(db *gorm.DB, num uint64) (BlockIntf, error)
	SetBlock(db *gorm.DB) error
}

// Block is the exported static model interface.
var Block block

// block ...
type block struct {
	Number    uint64 `gorm:"column:number;primary_key" json:"block_num"`
	Hash      string `gorm:"column:hash" json:"block_hash"`
	Time      uint64 `gorm:"column:time" json:"block_time"`
	Parent    string `gorm:"column:parent" json:"parent_hash"`
	Stable    bool   `gorm:"column:stable;defalt:false" json:"stable"`
	CreatedAt int64  `gorm:"column:created_at;default:extract(epoch from now())*1000" json:"-"`
	UpdatedAt int64  `gorm:"column:updated_at;default:extract(epoch from now())*1000" json:"-"`
}

func init() {
	registerModelForAutoMigration(&block{})
}

// TableName is used by GORM to choose which table to use.
func (b *block) TableName() string {
	return "blocks"
}

// createIndexes ...
func (b *block) createIndexes(db *gorm.DB) error {
	return nil
}

// createUniqueIndexes ...
func (b *block) createUniqueIndexes(db *gorm.DB) error {
	return nil
}

// createForeignKeys ...
func (b *block) createForeignKeys(db *gorm.DB) error {
	return nil
}

// GetNumber ...
func (b *block) GetNumber() uint64 {
	return b.Number
}

// GetHash ...
func (b *block) GetHash() string {
	return b.Hash
}

// GetTime ...
func (b *block) GetTime() uint64 {
	return b.Time
}

// GetParent ...
func (b *block) GetParent() string {
	return b.Parent
}

// GetStable ...
func (b *block) GetStable() bool {
	return b.Stable
}

// GetCreatedAt ...
func (b *block) GetCreatedAt() int64 {
	return b.CreatedAt
}

// GetUpdatedAt ...
func (b *block) GetUpdatedAt() int64 {
	return b.UpdatedAt
}

// SetBlocks ...
func (b *block) SetBlock(db *gorm.DB) error {
	return db.Where("number = ?", b.Number).FirstOrCreate(b).Error
}

// NewBlock
func NewBlock(ethBlock *types.Block) BlockIntf {
	newBlock := block{
		Number: ethBlock.Header().Number.Uint64(),
		Hash:   ethBlock.Hash().String(),
		Time:   ethBlock.Time(),
		Parent: ethBlock.Header().ParentHash.String(),
	}

	return &newBlock
}

// GetBlocks ...
func (b *block) GetBlocks(db *gorm.DB, n uint64) ([]BlockIntf, error) {
	// Get latest n blocks
	blocks := []*block{}
	err := db.Model(b).
		Order("number desc").
		Limit(n).
		Find(&blocks).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Organize into BlockIntf slice.
	blockIntfs := []BlockIntf{}
	for _, block := range blocks {
		blockIntfs = append(blockIntfs, block)
	}

	return blockIntfs, nil
}

// GetByNumber ...
func (b *block) GetByNumber(db *gorm.DB, num uint64) (BlockIntf, error) {
	// Get block based on given number
	block := block{}
	err := db.Model(b).Where("number = ?", num).First(&block).Error
	if err != nil {
		return nil, err
	}

	return &block, nil
}
