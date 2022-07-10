package api

import (
	"net/http"

	"main/api/middleware"
	"main/database"
	"main/models"

	"github.com/gin-gonic/gin"
)

func init() {
	// Setup domains router group.
	root := GetRoot().Group("transaction",
		middleware.FormatResponse())
	root.GET("/:txHash", GetByTxHash)
}

// GetByTxHash ...
func GetByTxHash(ctx *gin.Context) {
	// Get TxHash from URL path parameter.
	txHash := ctx.Param("txHash")

	// Get the trasaction by give txHash
	db := database.GetSQL()
	transaction, err := models.Transaction.GetByHash(db, txHash)
	if err != nil {
		respondWithErrorMessage(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	// Set results to context.
	ctx.Set("response", transaction)
}
