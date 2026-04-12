package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"

	"quriny.dev/internal/api"
	"quriny.dev/internal/database"
	"quriny.dev/internal/dsl"
	"quriny.dev/internal/ir"
	"quriny.dev/internal/middleware"
	"quriny.dev/internal/store"
)

// main bootstraps the Quriny runtime server:
//  1. Load and validate the DSL model from a JSON file.
//  2. Build the intermediate representation (IR) graph.
//  3. Initialise the chosen persistence store (in-memory or PostgreSQL).
//  4. Start the HTTP server with middleware (CORS, logging).
func main() {
	// CLI flags for server configuration.
	addr := flag.String("addr", ":8080", "HTTP listen address")
	modelPath := flag.String("model", "examples/product_app.json", "path to the app model JSON file")
	dbURL := flag.String("db", "", "PostgreSQL connection URL (leave empty for in-memory store)")
	flag.Parse()

	// --- Step 1: Load and validate the DSL model ---
	data, err := os.ReadFile(*modelPath)
	if err != nil {
		log.Fatalf("read model: %v", err)
	}

	model, err := dsl.LoadModel(data)
	if err != nil {
		log.Fatalf("load model: %v", err)
	}

	// --- Step 2: Build the runtime graph from the validated model ---
	graph, err := ir.BuildAppGraph(model)
	if err != nil {
		log.Fatalf("build app graph: %v", err)
	}

	// --- Step 3: Choose persistence backend based on flags ---
	var entityStore store.EntityStore
	if *dbURL != "" {
		// Connect to PostgreSQL and provision tables from the DSL model.
		ctx := context.Background()

		db, err := database.Connect(ctx, *dbURL)
		if err != nil {
			log.Fatalf("connect to database: %v", err)
		}
		defer db.Close()

		// Create tables for all entities defined in the DSL (idempotent).
		if err := db.ProvisionEntities(ctx, model); err != nil {
			log.Fatalf("provision entities: %v", err)
		}

		entityStore = store.NewPostgresStore(db.Pool, model)
		log.Println("Using PostgreSQL store")
	} else {
		entityStore = store.NewMemoryStore(model)
		log.Println("Using in-memory store (data will not persist across restarts)")
	}

	// --- Step 4: Build the HTTP handler stack ---
	server := api.NewServer(model, graph, entityStore)

	// Wrap the handler with middleware: outermost middleware runs first.
	handler := middleware.Chain(
		server.Handler(),
		middleware.RequestLogger, // log every request with method, path, status, duration
		middleware.CORS,          // allow cross-origin requests from the frontend dev server
	)

	log.Printf(
		"Quriny runtime listening on %s (model=%s, entities=%d, pages=%d, components=%d)",
		*addr,
		*modelPath,
		len(model.Entities),
		len(model.Pages),
		len(model.Components),
	)
	log.Printf("Available endpoints: %s, %s, %s, %s", "/health", "/app-model", "/app-graph", "/entities/{name}")

	if err := http.ListenAndServe(*addr, handler); err != nil {
		log.Fatalf("listen and serve: %v", err)
	}
}
