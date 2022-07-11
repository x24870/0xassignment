package eth_index

import (
	"fmt"
	"main/config"
	"main/database"
	"main/logging"
	"main/models"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/jinzhu/gorm"
	"golang.org/x/net/context"

	"github.com/ethereum/go-ethereum/ethclient"
)

// Static configuration variables initalized at runtime.
var comfirmedBlock uint64
var endpointURL string
var wsEndpointURL string

// eth index root context.
var ethRootCtx context.Context
var ethCancel context.CancelFunc

// init loads the logging configurations.
func init() {
	comfirmedBlock = config.GetUint64("COMFIRMED_BLOCK")
	endpointURL = config.GetString("INFURA_ENDPOINT")
	wsEndpointURL = config.GetString("INFURA_WS_ENDPOINT")
}

// Initialize initializes the logger module.
func Initialize(ctx context.Context) {
	// init root context and cancel function
	ethRootCtx, ethCancel = context.WithCancel(ctx)

	// connect to rpc endpoints and sync
	go subscribeAndSync(ethRootCtx, endpointURL, wsEndpointURL)
}

// Finalize finalizes the logging module.
func Finalize() {
	ethCancel()
}

// subscribeAndSync subscribe to ws endpoints and sync latest block to DB
func subscribeAndSync(ctx context.Context, endpointURL, wsEndpointURL string) {
	headers := make(chan *types.Header)
	defer close(headers)

	// connect to infura endpoint
	client, err := ethclient.Dial(endpointURL)
	if err != nil {
		logging.Error(ctx, err.Error())
	}
	defer client.Close()

	// connect to infura ws endpoint
	wsclient, err := ethclient.Dial(wsEndpointURL)
	if err != nil {
		logging.Error(ctx, err.Error())
	}
	defer wsclient.Close()

	// subscribe for new block head
	sub, err := wsclient.SubscribeNewHead(context.Background(), headers)
	if err != nil {
		logging.Error(ctx, err.Error())
	}
	defer sub.Unsubscribe()

	// continuously sync latest block to DB
SYNC:
	for {
		select {
		case err := <-sub.Err():
			logging.Error(ctx, err.Error())
		case header := <-headers:
			go getBlockAndSync(ctx, client, header.Number.Uint64(), false)
			go getBlockAndSync(ctx, client, header.Number.Uint64()-comfirmedBlock, true)
		case <-ctx.Done():
			logging.Info(ctx, "stop subscription")
			break SYNC
		}
	}
}

// getBlockAndSync get 1 block and sync block, transactions, receipt, logs to DB
func getBlockAndSync(
	ctx context.Context, client *ethclient.Client, blockNum uint64, stable bool) {
	blocksCh := getBlocks(ctx, client, []uint64{blockNum})

	block := <-blocksCh
	db := database.GetSQL()
	logging.Info(ctx, fmt.Sprintf("Real time Sync block: %d", block.Number().Uint64()))

	// check if the block in DB should be replace
	saveBlock := compareHashAndUpdate(ctx, db, block)
	if !saveBlock {
		return
	}

	// get block model instance and sync to DB
	newBlock := models.NewBlock(block)
	newBlock.SetStable(stable)
	if err := newBlock.SetBlock(db); err != nil {
		logging.Error(ctx, err.Error())
	}

	// sync transactions to DB
	blockHash := block.Header().Hash().String()
	saveTransactions(ctx, db, block.Transactions(), blockHash)

	// get receipts by tx hashes
	receiptCh := getReceipt(ctx, client, block.Transactions())
	for receipt := range receiptCh {
		// get receipt model instance and sync to DB
		newReceipt, err := models.NewReceipt(receipt)
		if err != nil {
			logging.Error(ctx, err.Error())
			continue
		}
		if err := newReceipt.SetReceipt(db); err != nil {
			logging.Error(ctx, err.Error())
		}

		saveTransactionLogs(ctx, db, receipt.Logs, receipt.TxHash.String())
	}
}

// getBlocks parallelly get blocks by provided block numbers
// return a channel contains blocks
func getBlocks(
	ctx context.Context, client *ethclient.Client, blockNums []uint64) chan *types.Block {
	ret := make(chan *types.Block, len(blockNums))
	wg := sync.WaitGroup{}
	wg.Add(len(blockNums))

	// get blocks by block number
	for _, num := range blockNums {
		go func(ctx context.Context, num uint64) {
			defer wg.Done()
			n := new(big.Int).SetUint64(num)
			block, err := client.BlockByNumber(ctx, n)
			if err != nil {
				logging.Error(ctx, err.Error())
				return
			}
			select {
			case ret <- block:
			case <-ctx.Done():
			}
		}(ctx, num)
	}

	// close channel if all goroutine done
	go func() {
		wg.Wait()
		close(ret)
	}()

	return ret
}

// getReceipt parallelly get receipts by provided tx hash
// return a channel contains receipts
func getReceipt(
	ctx context.Context, client *ethclient.Client, txHashes []*types.Transaction) chan *types.Receipt {
	ret := make(chan *types.Receipt, len(txHashes))
	wg := sync.WaitGroup{}
	wg.Add(len(txHashes))

	// get receipt by tx hash
	for _, txHash := range txHashes {
		go func(txHash common.Hash) {
			defer wg.Done()
			receipt, err := client.TransactionReceipt(ctx, txHash)
			if err != nil {
				logging.Error(ctx, err.Error())
				return
			}
			select {
			case ret <- receipt:
			case <-ctx.Done():
			}
		}(txHash.Hash())
	}

	// close channel if all goroutine done
	go func() {
		wg.Wait()
		close(ret)
	}()

	return ret
}

// compareHashAndUpdate compare the requested block hash with block hash in DB
// if the hashes are the same, update the block in DB to stable
// else delete the block data in DB
// returns if the requested block should be save
func compareHashAndUpdate(ctx context.Context, db *gorm.DB, block *types.Block) bool {
	old, err := models.Block.GetByNumber(db, block.Number().Uint64())
	// DB error, shouldn't save the block
	if err != nil && err != gorm.ErrRecordNotFound {
		logging.Error(ctx, err.Error())
		return false
	}
	// if hashes are the same, update the block to stable
	// and don't have to update other data
	if old != nil && old.GetHash() == block.Hash().String() {
		if err := old.UpdateBlockStable(db, true); err != nil {
			logging.Error(ctx, err.Error())
		}
		return false
	}
	// if hashes not matching, delete the block in the DB
	if old != nil && old.GetHash() != block.Hash().String() {
		logging.Info(ctx, fmt.Sprintf("delete: %d\n", old.GetNumber()))
		// delete fail, shouldn't save the block
		if err := old.DeleteBlock(db); err != nil {
			logging.Error(ctx, err.Error())
			return false
		}
		// delete successful, save this newer block
		return true
	}
	// old == nil, save this newer block
	return true
}

// SyncLastestBlocks Sync latest N blocks to DB
func SyncLastestBlocks(ctx context.Context) {
	logging.Info(ctx, fmt.Sprintf("Sync latest %d blocks...", comfirmedBlock))
	// init eth client
	client, err := ethclient.Dial(endpointURL)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	// get lastest N block numbers
	num, err := client.BlockNumber(ctx)
	if err != nil {
		logging.Critical(ctx, err.Error())
		return
	}

	logging.Info(ctx, fmt.Sprintf(
		"latest block num: %d\nSync blocks from %d to %d",
		num, num-comfirmedBlock, num))

	blockNums := make([]uint64, comfirmedBlock)
	for i := uint64(0); i < comfirmedBlock; i++ {
		blockNums[i] = num - i
	}

	// query latest N block parallelly
	blockCh := getBlocks(ctx, client, blockNums)

	// sync to DB ...
	db := database.GetSQL()
	for block := range blockCh {
		logging.Info(ctx, fmt.Sprintf("Sync block [%d]\n", block.Number().Uint64()))

		saveBlock := compareHashAndUpdate(ctx, db, block)
		if !saveBlock {
			continue
		}

		go func(ctx context.Context, db *gorm.DB, block *types.Block) {
			// get block model instance and sync to DB
			newBlock := models.NewBlock(block)
			if err := newBlock.SetBlock(db); err != nil {
				logging.Error(ctx, err.Error())
			}
			// sync transactions to DB
			blockHash := block.Header().Hash().String()
			saveTransactions(ctx, db, block.Transactions(), blockHash)

			// get receipts by tx hashes
			receiptCh := getReceipt(ctx, client, block.Transactions())
			for receipt := range receiptCh {
				// get receipt model instance and sync to DB
				newReceipt, err := models.NewReceipt(receipt)
				if err != nil {
					logging.Error(ctx, err.Error())
					continue
				}
				if err := newReceipt.SetReceipt(db); err != nil {
					logging.Error(ctx, err.Error())
				}

				saveTransactionLogs(ctx, db, receipt.Logs, receipt.TxHash.String())
			}
		}(ctx, db, block)

	}
}

func saveTransactions(
	ctx context.Context, db *gorm.DB, transactions []*types.Transaction, blockHash string) {
	for _, t := range transactions {
		newTransaction, err := models.NewTransaction(t, blockHash)
		if err != nil {
			logging.Error(ctx, err.Error())
			continue
		}
		if err := newTransaction.SetTransaction(db); err != nil {
			logging.Error(ctx, err.Error())
		}
	}
}

func saveTransactionLogs(
	ctx context.Context, db *gorm.DB, TxLogs []*types.Log, txHash string) {
	for _, txLog := range TxLogs {
		newTxLog, err := models.NewTransactionLog(txLog, txHash)
		if err != nil {
			logging.Error(ctx, err.Error())
			continue
		}
		if err := newTxLog.SetTransactionLog(db); err != nil {
			logging.Error(ctx, err.Error())
		}
	}
}
