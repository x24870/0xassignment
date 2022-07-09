package middleware

import (
	"net/http"

	"main/logging"

	"github.com/gin-gonic/gin"
)

// FormatResponse ...
func FormatResponse() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Onto the next handler if we're not final.
		ctx.Next()

		// Get response format.
		format := ctx.Query("format")
		if len(format) <= 0 {
			format = "json"
		}

		// Get prepared response from context.
		response, exists := ctx.Get("response")
		if !exists {
			errmsg := "no response"
			logging.Error(ctx.Request.Context(), errmsg)
			ctx.String(http.StatusInternalServerError, errmsg)
			return
		}

		// Respond based on specified format.
		switch format {
		case "yaml":
			ctx.YAML(http.StatusOK, response)
		case "json":
			ctx.JSON(http.StatusOK, response)
		}
	}
}
