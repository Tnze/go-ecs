package ecs

import "github.com/Tnze/go-ecs/internal/core"

type (
	World     = core.World
	Entity    = core.Entity
	Component = core.Component

	Filter      func(*World, *core.Archetype, *[]int) bool
	CachedQuery core.CachedQuery
)

// NewWorld creates a new empty World, with the default Components.
func NewWorld() (w *World) { return core.NewWorld() }

// NewEntity creates a new Entity in the World, without any Components.
func NewEntity(w *World) Entity { return core.NewEntity(w) }

// DelEntity deletes a Entity from the world.
func DelEntity(w *World, e Entity) { core.DelEntity(w, e) }

// NewComponent creates a new Component in the World.
// The data type associated with the Component will be bind when the first data is set.
func NewComponent(w *World) Component { return core.NewComponent(w) }

// AddComp adds the Component to Entity as a tag, without underlying content
func AddComp(w *World, e Entity, c Component) { core.AddComp(w, e, c) }

// HasComp reports whether the Entity has the Component.
func HasComp(w *World, e Entity, c Component) bool { return core.HasComp(w, e, c) }

// SetComp adds the Component and its content to Entity.
//
// If the Entity already has the Component, the content will be overridden.
// If the Entity doesn't have the Component, the Component will be added.
//
// This function panics if the type of data doesn't match others of the same Component.
func SetComp[C any](w *World, e Entity, c Component, data C) { core.SetComp[C](w, e, c, data) }

// GetComp gets the data of a Component of an Entity.
// If the Entity doesn't have the Component, nil will be returned.
func GetComp[C any](w *World, e Entity, c Component) (data *C) { return core.GetComp[C](w, e, c) }

// DelComp removes the Component of an Entity.
// If the Entity doesn't have the Component, nothing will happen.
func DelComp(w *World, e Entity, c Component) { core.DelComp(w, e, c) }

// QueryAll return a filter querying Entities that have all Components required.
func QueryAll(comps ...Component) Filter { return Filter(core.QueryAll(comps...)) }

// QueryAny return a filter querying Entities that have at least one of the required Component.
func QueryAny(comps ...Component) Filter { return Filter(core.QueryAny(comps...)) }

func (f Filter) Run(w *World, h func([]Entity, []any)) { core.Filter(f).Run(w, h) }
func (f Filter) Cache(w *World) *CachedQuery           { return (*CachedQuery)(core.Filter(f).Cache(w)) }
func (q *CachedQuery) Run(h func([]Entity, []any))     { (*core.CachedQuery)(q).Run(h) }
func (q *CachedQuery) Free(w *World)                   { (*core.CachedQuery)(q).Free(w) }

// debug

// Type is for debug purpose. Return a human-readable string representing the Archetype of an Entity.
//
// To use this function, make sure you have add Component[string] for your other Component
// (Component is valid Entity, you can add Component to them). Otherwise, the type name of the data is used.
func Type(w *World, e Entity, nameComp Component) string { return core.Type(w, e, nameComp) }
