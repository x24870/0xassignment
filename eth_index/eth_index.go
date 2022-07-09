package eth_index

import (
	"fmt"
	"main/database"
	"main/logging"
	"main/models"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/core/types"
	"golang.org/x/net/context"

	"github.com/ethereum/go-ethereum/ethclient"
)

// Static configuration variables initalized at runtime.
var comfirmedBlock uint64

// eth index root context.
var ethRootCtx context.Context

// init loads the logging configurations.
func init() {
	// comfirmedBlock = config.GetUint("COMFIRMED_BLOCK")
	comfirmedBlock = 1 //20
}

// Initialize initializes the logger module.
func Initialize(ctx context.Context) {
	// Save database root context.
	ethRootCtx = ctx

	// Setup timeout context for connecting to StackDriver.
	// timeoutCtx, cancel := context.WithTimeout(ctx, gRPCConnectTimeout)
	// defer cancel()

}

// getBlocks get blocks by provided block numbers parallelly
// return a channel contains block data
func getBlocks(ctx context.Context, client *ethclient.Client, blockNums []uint64) chan *types.Block {
	ret := make(chan *types.Block, len(blockNums))
	wg := sync.WaitGroup{}
	wg.Add(len(blockNums))

	// get blocks by block number
	for _, num := range blockNums {
		go func(num uint64) {
			n := new(big.Int).SetUint64(num)
			block, err := client.BlockByNumber(ctx, n)
			if err != nil {
				logging.Error(ctx, err.Error())
				return
			}
			select {
			case ret <- block:
				wg.Done()
			case <-ctx.Done():
				wg.Done()
			}
		}(num)
	}

	// close channel if all goroutine done
	go func() {
		wg.Wait()
		close(ret)
	}()

	return ret
}

func SyncLastestBlocks(ctx context.Context) {
	logging.Info(ctx, fmt.Sprintf("Sync latest %d blocks...", comfirmedBlock))
	// init eth client
	client, err := ethclient.Dial("https://data-seed-prebsc-2-s3.binance.org:8545/")
	if err != nil {
		panic(err)
	}
	defer client.Close()

	// get lastest N block numbers
	num, err := client.BlockNumber(ctx)
	logging.Info(ctx, fmt.Sprintf("latest block num: %d", num))
	if err != nil {
		logging.Error(ctx, err.Error())
	}

	blockNums := make([]uint64, comfirmedBlock)
	fmt.Println("blockNums: ", blockNums)
	for i := uint64(0); i < comfirmedBlock; i++ {
		blockNums[i] = num - i
		fmt.Println("i: ", i, "blockNums: ", blockNums)
	}

	// query latest N block parallelly
	out := getBlocks(ctx, client, blockNums)

	// sync to DB ...
	// db := database.GetSQL()
	db := database.GetSQL()
	for block := range out {
		fmt.Println(block.Header().Number, block.Header().Hash())
		newBlock := models.NewBlock(block)
		if err := newBlock.SetBlock(db); err != nil {
			logging.Error(ctx, err.Error())
		}

	}
}

// Finalize finalizes the logging module.
func Finalize() {

}
