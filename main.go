package main

import (
	"context"
	"fmt"
	"main/config"
	"main/database"
	"main/eth_index"
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

	// Setup database module
	database.Initialize(ctx)
	defer database.Finalize()

	// Setup etn_index module
	eth_index.Initialize(ctx)
	defer eth_index.Finalize()
	eth_index.SyncLastestBlocks(ctx)

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
