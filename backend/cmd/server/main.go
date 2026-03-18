package main

import (
	"fmt"
	"os"

	"quriny.dev/internal/dsl"
	"quriny.dev/internal/ir"
)

// main is a temporary bootstrap used to prove the current backend flow:
// load DSL JSON, validate it, then build the runtime graph.
func main() {
	data, err := os.ReadFile("examples/product_app.json")
	if err != nil {
		panic(err)
	}

	model, err := dsl.LoadModel(data)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Loaded model: %+v\n", model)

	fmt.Println()

	graph, err := ir.BuildAppGraph(model)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Loaded graph: %+v\n", graph)

	// http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
	// 	fmt.Fprintf(w, "Quriny Runtime OK")
	// })

	// http.ListenAndServe(":8080", nil)
}
