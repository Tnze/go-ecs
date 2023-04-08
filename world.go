package ecs

import (
	"hash/maphash"
	"reflect"
	"sort"
	"unsafe"
)

type (
	// The World is the container for all ECS data.
	// It stores the entities and their components, does queries and runs systems.
	//
	// --flecs.dev
	World struct {
		idManager
		zero *archetype

		entities   map[Entity]*entityRecord
		archetypes map[uint64]*archetype
		components map[Component]map[*archetype]int

		// internal Components
		NameComp Component
	}

	// An Entity is a unique thing in the world, and is represented by a 64-bit id.
	// Entities can be created and deleted.
	// If an entity is deleted, it is no longer considered "alive".
	//
	// A world can contain up to 4 billion alive entities.
	// Entity identifiers contain a few bits that make it possible to check whether an entity is alive or not.
	//
	// --flecs.dev
	Entity uint64

	// A Component is a type of which instances can be added and removed to entities.
	// Each component can be added only once to an entity.
	//
	// --flecs.dev
	Component struct{ Entity }

	idManager struct {
		nextID uint64
	}
	entityRecord struct {
		at  *archetype
		row int
	}
	archetype struct {
		types
		entities columnImpl[Entity]
		records  columnImpl[*entityRecord]
		comps    []column
		edges    map[Component]archetypeEdge
	}
	types         []componentMeta
	componentMeta struct {
		Component
		columnType reflect.Type
	}
	archetypeEdge struct {
		add, del *archetype
	}
	column interface {
		appendFrom(other column, column int)
		swapDelete(i int)
		len() int
	}
	columnImpl[C any] []C
)

// NewWorld creates a new empty World, with the default components.
func NewWorld() (w *World) {
	w = &World{
		entities:   make(map[Entity]*entityRecord),
		archetypes: make(map[uint64]*archetype),
		components: make(map[Component]map[*archetype]int),
	}
	w.zero = newArchetype(w, nil)

	// general components
	w.NameComp = NewComponent(w)
	Set(w, w.NameComp.Entity, w.NameComp, "ecs.Name")
	return
}

// NewEntity creates a new Entity in the World, without any components.
func NewEntity(w *World) (e Entity) {
	e = Entity(w.genID())
	r := new(entityRecord)
	r.at = w.zero
	r.row = w.zero.entities.append(e)
	w.zero.records.append(r)
	w.entities[e] = r
	return
}

// NewNamedEntity is the same as NewEntity, but automatically sets a name component.
func NewNamedEntity(w *World, name string) (e Entity) {
	e = NewEntity(w)
	Set(w, e, w.NameComp, name)
	return
}

// NewComponent creates a new Component in the World.
// The data type associated with the Component will be bind when the first data is set.
func NewComponent(w *World) (c Component) {
	c = Component{NewEntity(w)}
	w.components[c] = make(map[*archetype]int)
	return
}

// NewNamedComponent is the same as NewComponent, but automatically sets a name component.
func NewNamedComponent(w *World, name string) (c Component) {
	c = NewComponent(w)
	Set(w, c.Entity, w.NameComp, name)
	return
}

func newArchetype(w *World, t types) (a *archetype) {
	a = &archetype{
		types: t,
		comps: make([]column, len(t)),
		edges: make(map[Component]archetypeEdge),
	}
	for i, v := range t {
		a.comps[i] = reflect.New(v.columnType.Elem()).Interface().(column)
		w.components[v.Component][a] = i
	}
	w.archetypes[t.sortHash()] = a
	return
}

// Set sets the data of a Component of an Entity.
//
// If the Entity already has the Component, the data will be overridden.
// If the Entity doesn't have the Component, the Component will be added.
//
// This function panics if the type of data doesn't match others of the same Component.
func Set[C any](w *World, e Entity, c Component, data C) {
	rec := w.entities[e]
	// If the archetype of e already contains c.
	// Override the data and return.
	if col, ok := w.components[c][rec.at]; ok {
		(*rec.at.comps[col].(*columnImpl[C]))[rec.row] = data
		return
	}
	// Lookup archetypeEdge for shortcuts
	edge := rec.at.edges[c]
	target := edge.add
	if target == nil {
		// We don't have shortcuts yet. Use the hash way.
		var tmpS *columnImpl[C]
		var ok bool
		newTypes := rec.at.types.copyAppend(c, reflect.TypeOf(tmpS))
		if target, ok = w.archetypes[newTypes.sortHash()]; !ok {
			target = newArchetype(w, newTypes)
		}
		// Save to the shortcuts
		edge.add = target
		rec.at.edges[c] = edge
	}
	// Move entity to the new archetype
	row := target.entities.append(e)
	target.records.append(rec)
	rec.at.entities.swapDelete(rec.row)
	rec.at.records.swapDelete(rec.row)
	if rec.row != len(rec.at.entities) {
		w.entities[rec.at.entities[rec.row]].row = rec.row
	}
	target.comps[w.components[c][target]].(*columnImpl[C]).append(data)
	// Move other components
	for _, t := range rec.at.types {
		idx := w.components[t.Component]
		src := rec.at.comps[idx[rec.at]]
		target.comps[idx[target]].appendFrom(src, rec.row)
		src.swapDelete(rec.row)
	}

	rec.at = target
	rec.row = row
}

// Remove removes the Component of an Entity.
// If the Entity doesn't have the Component, nothing will happen.
func Remove(w *World, e Entity, c Component) {
	rec := w.entities[e]
	col, ok := w.components[c][rec.at]
	if !ok {
		return // archetype of e doesn't contain component c
	}
	// Lookup archetypeEdge for shortcuts
	edge := rec.at.edges[c]
	target := edge.del
	if target == nil {
		// We don't have shortcuts yet. Use the hash way.
		newTypes := rec.at.types.copyDelete(col)
		if target, ok = w.archetypes[newTypes.sortHash()]; !ok {
			target = newArchetype(w, newTypes)
		}
		// Save to the shortcuts
		edge.del = target
		rec.at.edges[c] = edge
	}
	// Move entity
	row := target.entities.append(e)
	target.records.append(rec)
	rec.at.entities.swapDelete(rec.row)
	rec.at.records.swapDelete(rec.row)
	if rec.row != len(rec.at.entities) {
		w.entities[rec.at.entities[rec.row]].row = rec.row
	}
	for _, t := range target.types {
		// Move other components
		idx := w.components[t.Component]
		src := rec.at.comps[idx[rec.at]]
		target.comps[idx[target]].appendFrom(src, rec.row)
		src.swapDelete(rec.row)
	}

	rec.at = target
	rec.row = row
}

// Get gets the data of a Component of an Entity.
// If the Entity doesn't have the Component, nil will be returned.
func Get[C any](w *World, e Entity, c Component) (data *C) {
	rec := w.entities[e]
	if column, ok := w.components[c][rec.at]; ok {
		return &(*rec.at.comps[column].(*columnImpl[C]))[rec.row]
	}
	return nil
}

func (i *idManager) genID() (id uint64) {
	id = i.nextID
	i.nextID++
	return
}

func (t types) sortHash() uint64 {
	// sort the component list
	sort.Slice(t, func(i, j int) bool {
		return t[i].Component.Entity < t[i].Component.Entity
	})
	// calculate hash
	var h maphash.Hash
	for i := range t {
		h.Write((*[8]byte)(unsafe.Pointer(&t[i].Component))[:])
	}
	return h.Sum64()
}

func (t types) copyAppend(c Component, storeType reflect.Type) (newTypes types) {
	newTypes = make(types, len(t)+1)
	newTypes[0] = componentMeta{
		Component:  c,
		columnType: storeType,
	}
	copy(newTypes[1:], t)
	return
}

func (t types) copyDelete(i int) (newTypes types) {
	newTypes = make(types, len(t)-1)
	copy(newTypes[:i], t[:i])
	if i+1 < len(t) {
		copy(newTypes[i:], t[i+1:])
	}
	return
}

func (c columnImpl[C]) len() int {
	return len(c)
}

func (c *columnImpl[C]) appendFrom(other column, row int) {
	*c = append(*c, (*other.(*columnImpl[C]))[row])
}

func (c *columnImpl[C]) swapDelete(i int) {
	last := len(*c) - 1
	(*c)[i] = (*c)[last]
	*c = (*c)[:last]
}

func (c *columnImpl[C]) append(data C) int {
	*c = append(*c, data)
	return len(*c) - 1
}
