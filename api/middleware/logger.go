package middleware

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"main/logging"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// blacklist is a list of request URLs that we should ignore from logging.
var blacklist = map[string]bool{
	"/alive": false,
	"/ready": false,
}

// Logger returns a request logger middleware, which logs the HTTP request and
// creates a logger instance to be used throughout the execution of the request.
func Logger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Do nothing if the request URL is on the blacklist.
		url := ctx.Request.URL.EscapedPath()
		if _, exists := blacklist[url]; exists {
			return
		}

		// Inject request ID into context.Context of *http.Request of *gin.Context
		requestID := generateRequestID(ctx.Request)
		ctxWithRequestID := context.WithValue(ctx.Request.Context(),
			logging.ContextKeyRequestID, requestID)
		ctx.Request = ctx.Request.WithContext(ctxWithRequestID)

		// Collect relevant information from this request to be logged.
		address := ctx.ClientIP()
		method := ctx.Request.Method
		params := ctx.Request.URL.RawQuery
		headersMap, err := json.Marshal(ctx.Request.Header)
		if err != nil {
			logging.Error(ctx.Request.Context(),
				"Failed to marshal headers: %v", err)
			headersMap = []byte{}
		}
		headers := string(headersMap)

		// Log the incoming request information.
		logging.Info(ctx.Request.Context(), "Client: [%15s], Method: [%6s], "+
			"Path: [%s], Params: [%s], Headers: %s", address, method, url,
			params, headers)

		// Continue processing request chain while measuring response time.
		start := time.Now()
		ctx.Next()
		elapsed := time.Since(start)

		// Get response code.
		code := ctx.Writer.Status()

		// Log the request body on error.
		var body string
		if (method == http.MethodPost || method == http.MethodPatch) &&
			code >= http.StatusBadRequest {
			body = string(GetBody(ctx))
		}

		// Log the outgoing response information.
		logging.Info(ctx.Request.Context(), "Code: [%3d], Latency: [%10v], "+
			"Body: [%s] , Path: [%s] ", code, elapsed, body, url)

	}
}

// GetRequestID returns the request ID associated with the current request.
func GetRequestID(ctx *gin.Context) string {
	// Lookup the request logger.
	requestID, ok := ctx.Request.Context().Value(logging.ContextKeyUserID).(string)
	if !ok {
		logging.Error(ctx.Request.Context(), "Failed to lookup request ID")
		return ""
	}

	return requestID
}

func generateRequestID(request *http.Request) string {
	// Generate hash object.
	hash := fnv.New64a()

	// Use time as hash component.
	currentTimeBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(currentTimeBytes,
		uint64(time.Now().UnixNano()))

	// Compute hash value.
	hash.Write([]byte(request.Host))
	hash.Write([]byte(request.RemoteAddr))
	hash.Write([]byte(request.RequestURI))
	hash.Write(currentTimeBytes)

	return fmt.Sprintf("%012x", hash.Sum64())[:12]
}
