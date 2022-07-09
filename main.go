package main

import (
	"context"
	"fmt"
	"main/config"
	"main/global"
	"main/logging"
	"main/server"
	_ "net/http/pprof"
)

func main() {
	// Prepare readiness & liveness flags.
	global.Ready = false
	global.Alive = false

	// Create root context.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup logging module.
	// NOTE: This should always be first.
	logging.Initialize(ctx)
	defer logging.Finalize()

	// Create HTTP server instance to listen on all interfaces.
	address := fmt.Sprintf("%s:%s",
		config.GetString("SERVER_LISTEN_ADDRESS"),
		config.GetString("SERVER_LISTEN_PORT"))
	server := server.CreateServer(ctx, address)

	// Set readiness & liveness flags.
	global.Ready = true
	global.Alive = true

	// Start servicing requests.
	logging.Info(ctx, "Initialization complete, listening on %s...", address)
	if err := server.ListenAndServe(); err != nil {
		logging.Info(ctx, err.Error())
	}
}