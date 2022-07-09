package server

import (
	"context"
	"main/api"
	"main/config"
	"main/global"
	"main/logging"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

// CreateServer creates an HTTP server listening on the specified address.
func CreateServer(ctx context.Context, address string) *http.Server {
	// Start pprof server if in debug mode.
	go func() {
		http.ListenAndServe("127.0.0.1:9453", nil)
	}()

	// Setup HTTP Server.
	server := &http.Server{
		Addr:    address,
		Handler: api.GetRouter(),
	}

	// Install the shutdown handler.
	installShutdownHandler(ctx, server)

	return server
}

// installShutdownHandler registers a shutdown handler for graceful shutdown.
func installShutdownHandler(ctx context.Context, server *http.Server) {
	// Create signal channel & shutdown timeout context.
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	timeoutCtx, cancel := context.WithTimeout(ctx,
		config.GetMilliseconds("SERVER_SHUTDOWN_GRACE_PERIOD_MS"))

	// Catch signals in a separate goroutine.
	go func(ctx context.Context) {
		defer cancel()

		// Wait for signals.
		sig := <-sigChan
		signal.Stop(sigChan)
		logging.Warn(ctx, "Received signal: %s.", sig.String())

		// Perform graceful shutdown.
		logging.Warn(ctx, "Initiating graceful shutdown...")
		global.Alive = false
		if err := server.Shutdown(timeoutCtx); err != nil {
			logging.Error(ctx, "Failed to shutdown: %s", err.Error())
		}
	}(ctx)
}
