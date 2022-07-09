package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"

	"main/logging"
)

// Recovery is a middleware that recovers from panic then logs the stack trace.
func Recovery() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			// Recover from panic.
			if recovered := recover(); recovered != nil {
				// Assemble log string.
				message := fmt.Sprintf("\x1b[31m%v\n[Stack Trace]\n%s\x1b[m",
					recovered, debug.Stack())

				// Record the stack trace to logging service.
				logging.Error(ctx.Request.Context(), message)

				// Discontinue the request handler chain processing.
				ctx.AbortWithStatus(http.StatusInternalServerError)
			}
		}()

		// Continue processing request chain.
		ctx.Next()
	}
}
