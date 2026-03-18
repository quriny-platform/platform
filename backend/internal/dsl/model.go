// Package dsl defines the authoring model used to describe a Quriny app
// before it is validated and transformed into the internal graph/IR form.
package dsl

// AppModel is the top-level DSL document for a Quriny application.
type AppModel struct {
	Entities   []Entity     `json:"entities"`
	Pages      []Page       `json:"pages"`
	Components []Component  `json:"components"`
	Actions    []Action     `json:"actions"`
	Navigation []Navigation `json:"navigation"`
}

// Entity describes a business object exposed by the application model.
type Entity struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Fields []Field `json:"fields"`
}

// Field defines a single property on an entity.
type Field struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// Page describes a route and the components rendered on that route.
type Page struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Path       string   `json:"path"`
	Components []string `json:"components"`
}

// Component describes a UI building block bound to data and actions.
type Component struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Type    string   `json:"type"`
	Entity  string   `json:"entity"`
	Fields  []string `json:"fields"`
	Actions []string `json:"actions"`
}

// Action describes an operation that can be triggered from the UI or runtime.
type Action struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Entity string `json:"entity"`
}

// Navigation describes how a page is exposed in the app's navigation model.
type Navigation struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Path  string `json:"path"`
	Page  string `json:"page"`
}
