// Package ir defines the intermediate representation used after a DSL model is
// validated. The IR is graph-shaped so the runtime can reason about relations
// between pages, components, entities, and actions more directly.
package ir

// NodeID is the stable identifier used inside the application graph.
type NodeID string

// NodeType identifies what kind of runtime object a node represents.
type NodeType string

const (
	NodePage      NodeType = "page"
	NodeComponent NodeType = "component"
	NodeEntity    NodeType = "entity"
	NodeAction    NodeType = "action"
)

// Node stores one model element plus its incoming and outgoing relationships.
type Node struct {
	ID   NodeID
	Type NodeType

	// Data keeps the original DSL value attached to the graph node.
	Data any

	Incoming []*Edge
	Outgoing []*Edge
}

// Edge expresses a directed relationship between two graph nodes.
type Edge struct {
	From NodeID
	To   NodeID
	Type string
}

// AppGraph is the runtime-friendly representation derived from the DSL.
type AppGraph struct {
	Nodes map[NodeID]*Node
	Edges []*Edge
}
