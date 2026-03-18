package ir

import (
	"fmt"

	"quriny.dev/internal/dsl"
)

// BuildAppGraph converts the authoring DSL model into the runtime graph/IR.
func BuildAppGraph(model *dsl.AppModel) (*AppGraph, error) {
	g := &AppGraph{
		Nodes: make(map[NodeID]*Node),
	}

	// 1. Create Entity nodes
	for _, e := range model.Entities {
		id := NodeID("entity:" + e.Name)

		g.Nodes[id] = &Node{
			ID:   id,
			Type: NodeEntity,
			Data: e,
		}
	}

	// 2. Create Component nodes
	for _, c := range model.Components {
		id := NodeID("component:" + c.Name)

		g.Nodes[id] = &Node{
			ID:   id,
			Type: NodeComponent,
			Data: c,
		}
	}

	// 3. Create Page nodes
	for _, p := range model.Pages {
		id := NodeID("page:" + p.Name)

		g.Nodes[id] = &Node{
			ID:   id,
			Type: NodePage,
			Data: p,
		}
	}

	// 4. Create Action nodes
	for _, a := range model.Actions {
		id := NodeID("action:" + a.Name)

		g.Nodes[id] = &Node{
			ID:   id,
			Type: NodeAction,
			Data: a,
		}
	}

	// --- EDGES ---

	// Page → Components
	for _, p := range model.Pages {
		from := NodeID("page:" + p.Name)

		for _, compName := range p.Components {
			to := NodeID("component:" + compName)

			if err := connect(g, from, to, "renders"); err != nil {
				return nil, err
			}
		}
	}

	// Component → Entity
	for _, c := range model.Components {
		if c.Entity == "" {
			continue
		}

		from := NodeID("component:" + c.Name)
		to := NodeID("entity:" + c.Entity)

		if err := connect(g, from, to, "uses"); err != nil {
			return nil, err
		}
	}

	// Component → Actions
	for _, c := range model.Components {
		from := NodeID("component:" + c.Name)

		for _, act := range c.Actions {
			to := NodeID("action:" + act)

			if err := connect(g, from, to, "triggers"); err != nil {
				return nil, err
			}
		}
	}

	return g, nil
}

// connect creates a typed edge and keeps both adjacency lists in sync.
func connect(g *AppGraph, from, to NodeID, edgeType string) error {
	fromNode, ok := g.Nodes[from]
	if !ok {
		return fmt.Errorf("unknown node: %s", from)
	}

	toNode, ok := g.Nodes[to]
	if !ok {
		return fmt.Errorf("unknown node: %s", to)
	}

	e := &Edge{
		From: from,
		To:   to,
		Type: edgeType,
	}

	fromNode.Outgoing = append(fromNode.Outgoing, e)
	toNode.Incoming = append(toNode.Incoming, e)
	g.Edges = append(g.Edges, e)

	return nil
}
