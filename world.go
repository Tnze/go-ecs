package ecs

import (
	"hash/maphash"
	"reflect"
	"sort"
	"unsafe"
)

type (
	World struct {
		idManager
		zero *archetype

		entities   map[Entity]*entityRecord
		archetypes map[uint64]*archetype
		components map[Component]map[*archetype]int

		// internal Components
		NameComp Component
	}
	Entity    uint64
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

	// general components
)

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

func NewEntity(w *World) (e Entity) {
	e = Entity(w.genID())
	r := new(entityRecord)
	r.at = w.zero
	r.row = w.zero.entities.append(e)
	w.zero.records.append(r)
	w.entities[e] = r
	return
}

func NewNamedEntity(w *World, name string) (e Entity) {
	e = NewEntity(w)
	Set(w, e, w.NameComp, name)
	return
}

func NewComponent(w *World) (c Component) {
	c = Component{NewEntity(w)}
	w.components[c] = make(map[*archetype]int)
	return
}

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
	w.entities[e].row = rec.row
	target.comps[w.components[c][target]].(*columnImpl[C]).append(data)
	for _, t := range rec.at.types {
		// Move other components
		idx := w.components[t.Component]
		src := rec.at.comps[idx[rec.at]]
		target.comps[idx[target]].appendFrom(src, rec.row)
		src.swapDelete(rec.row)
	}

	rec.at = target
	rec.row = row
}

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
	w.entities[e].row = rec.row
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
	// sort the components list
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
