package api

import (
	"net/http"
	"strconv"

	"main/api/middleware"
	"main/config"
	"main/database"
	"main/models"

	"github.com/gin-gonic/gin"
)

type BlockWithTransaction struct {
	Block        models.BlockIntf
	Transactions []string
}

func init() {
	// Setup domains router group.
	root := GetRoot().Group("blocks",
		middleware.FormatResponse())
	root.GET("", GetBlocks)
	root.GET("/:id", GetBlockByNumber)
}

// GetBlocks ...
func GetBlocks(ctx *gin.Context) {
	// Get the number of requested blocks
	num, err := strconv.ParseUint(ctx.Query("limit"), 10, 64)
	if err != nil {
		respondWithErrorMessage(ctx, http.StatusBadRequest, "invalid limit")
		return
	}
	if num > config.GetUint64("API_MAX_BLOCK_REQ") {
		respondWithErrorMessage(ctx, http.StatusBadRequest, "require too many blocks")
		return
	}

	// Get latest N blocks
	db := database.GetSQL()
	blocks, err := models.Block.GetBlocks(db, num)
	if err != nil {
		respondWithErrorMessage(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	// Set results to context.
	ctx.Set("response", blocks)
}

// GetBlockByNumber ...
func GetBlockByNumber(ctx *gin.Context) {
	// Get Block Number from URL path parameter.
	num, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		respondWithErrorMessage(ctx, http.StatusBadRequest, "invalid block ID")
		return
	}

	// Get the block by given block number
	db := database.GetSQL()
	block, err := models.Block.GetByNumber(db, num)
	if err != nil {
		respondWithErrorMessage(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	// Get transactions in the block
	transactions, err := models.Transaction.GetByBlockHash(db, block.GetHash())
	if err != nil {
		respondWithErrorMessage(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	// Organize block and transactions
	txHashes := make([]string, len(transactions))
	for i, v := range transactions {
		txHashes[i] = v.GetTxHash()
	}
	resp := BlockWithTransaction{Block: block, Transactions: txHashes}

	// Set results to context.
	ctx.Set("response", resp)
}
