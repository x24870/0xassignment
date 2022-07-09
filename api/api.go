package api

import (
	"fmt"
	"main/api/middleware"
	"main/config"
	"main/global"
	"main/logging"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

// The global HTTP router instance and root group.
var router *gin.Engine
var root *gin.RouterGroup
var once sync.Once

// respondWithErrorMessage responds to the request with the provided error message.
func respondWithErrorMessage(ctx *gin.Context, status int,
	format string, args ...interface{}) {

	// Compose error message for logging and response.
	message := fmt.Sprintf(format, args...)

	// Log error message and respond to request.
	logging.Error(ctx.Request.Context(), message)
	ctx.AbortWithStatusJSON(status, gin.H{
		"error":      message,
		"request_id": middleware.GetRequestID(ctx),
	})
}

// GetRouter returns the global HTTP router instance.
func GetRouter() *gin.Engine {
	// Initialize API singleton instances.
	once.Do(initializeSingletons)
	return router
}

// GetRoot returns the router root group.
func GetRoot() *gin.RouterGroup {
	// Initialize API singleton instances.
	once.Do(initializeSingletons)
	return root
}

// initializeSingletons is the function called by sync.Once to intialize the
// HTTP engine and router group singleton instances.
func initializeSingletons() {
	// Create router and group instances. Check whether we should use the
	// microservice name as root router group URL prefix. This depends on
	// whether or our Kubernetes ingress is configured to use path-based
	// routing or name-based virtual hosting.
	if config.GetBool("SERVICE_NAME_AS_ROOT") {
		router, root = createRouterAndGroup(global.ServiceName)
	} else {
		router, root = createRouterAndGroup("")
	}
}

// Create a clean router and a root group with the given microservice prefix.
func createRouterAndGroup(prefix string) (*gin.Engine, *gin.RouterGroup) {
	// Create a clean HTTP router engine.
	engine := gin.New()

	// Configure HTTP router engine settings.
	engine.RedirectTrailingSlash = true
	engine.RedirectFixedPath = false
	engine.HandleMethodNotAllowed = true
	engine.ForwardedByClientIP = true

	// Create from the engine a router group with the given prefix.
	group := engine.Group(prefix)

	// Install common middleware to the router group.
	installCommonMiddleware(group)

	// Install default not found handler.
	installNotFoundHandler(engine)

	return engine, group
}

// installCommonMiddleware installs common middleware to the router group.
func installCommonMiddleware(group *gin.RouterGroup) {
	// Install logger middleware, a middleware to log requests.
	group.Use(middleware.Logger())

	// Install recovery middleware, a middleware to recover & log panics.
	// NOTE: The recovery middleware should always be the last one installed.
	group.Use(middleware.Recovery())
}

// installNotFoundHandler installs default handler for unmatched routes.
func installNotFoundHandler(engine *gin.Engine) {
	engine.NoRoute(func(ctx *gin.Context) {
		ctx.String(http.StatusNotFound, "Sorry, this endpoint does not exist.\n")
	})
}
