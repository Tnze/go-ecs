package ecs

import (
	"hash/maphash"
	"reflect"
	"sort"
	"strings"
	"unsafe"
)

type (
	World struct {
		idManager
		zero *archetype

		entities   map[Entity]entityRecord
		archetypes map[uint64]*archetype
		components map[Component]map[*archetype]int
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
		comps []store
		edges map[Component]archetypeEdge
	}
	types         []componentMeta
	componentMeta struct {
		Component
		name      string
		storeType reflect.Type
	}
	archetypeEdge struct {
		add, del *archetype
	}
	store interface {
		appendFrom(other store, column int)
		swapDelete(i int)
		len() int
	}
	storeImpl[C any] []C
)

func NewWorld() (w *World) {
	w = &World{
		entities:   make(map[Entity]entityRecord),
		archetypes: make(map[uint64]*archetype),
		components: make(map[Component]map[*archetype]int),
	}
	w.zero = newArchetype(w, nil)
	return
}

func NewEntity(w *World) (e Entity) {
	e = Entity(w.genID())
	w.entities[e] = entityRecord{
		at:  w.zero,
		row: -1,
	}
	return
}

func NewComponent(w *World) (c Component) {
	c = Component{NewEntity(w)}
	w.components[c] = make(map[*archetype]int)
	return
}

func newArchetype(w *World, t types) (a *archetype) {
	a = &archetype{
		types: t,
		comps: make([]store, len(t)),
		edges: make(map[Component]archetypeEdge),
	}
	for i, v := range t {
		a.comps[i] = reflect.New(v.storeType.Elem()).Interface().(store)
		w.components[v.Component][a] = i
	}
	w.archetypes[t.sortHash()] = a
	return
}

func Set[C any](w *World, e Entity, c Component, data C) {
	rec := w.entities[e]
	// If the archetype of e already contains c.
	// Override the data and return.
	if column, ok := w.components[c][rec.at]; ok {
		(*rec.at.comps[column].(*storeImpl[C]))[rec.row] = data
		return
	}
	// Lookup archetypeEdge for shortcuts
	edge := rec.at.edges[c]
	target := edge.add
	if target == nil {
		// We don't have shortcuts yet. Use the hash way.
		var tmpC *C
		var tmpS *storeImpl[C]
		newTypes := rec.at.types.copyAppend(c, reflect.TypeOf(tmpS), reflect.TypeOf(tmpC).Elem().Name())
		target = newArchetype(w, newTypes)
		// Save to the shortcuts
		edge.add = target
		rec.at.edges[c] = edge
	}
	// Move entity to the new archetype
	row := target.comps[w.components[c][target]].(*storeImpl[C]).append(data)
	for _, t := range rec.at.types {
		// copy
		idx := w.components[t.Component]
		src := rec.at.comps[idx[rec.at]]
		target.comps[idx[target]].appendFrom(src, rec.row)
	}

	w.entities[e] = entityRecord{at: target, row: row}
}

func Remove[C any](w *World, e Entity, c Component) {}

func Get[C any](w *World, e Entity, c Component) (data *C) {
	rec := w.entities[e]
	if column, ok := w.components[c][rec.at]; ok {
		return &(*rec.at.comps[column].(*storeImpl[C]))[rec.row]
	}
	return nil
}

func Type(w *World, e Entity) string {
	var sb strings.Builder
	rec := w.entities[e]
	compNames := make([]string, len(rec.at.types))
	for i, v := range rec.at.types {
		compNames[i] = v.name
	}
	sort.Strings(compNames)
	for i, v := range compNames {
		if i != 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(v)
	}
	return sb.String()
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

func (t types) copyAppend(c Component, storeType reflect.Type, name string) (newTypes types) {
	newTypes = make(types, len(t)+1)
	newTypes[0] = componentMeta{
		Component: c,
		name:      name,
		storeType: storeType,
	}
	copy(newTypes[1:], t)
	return
}

func (s storeImpl[C]) len() int {
	return len(s)
}

func (s *storeImpl[C]) appendFrom(other store, row int) {
	*s = append(*s, (*other.(*storeImpl[C]))[row])
}

func (s *storeImpl[C]) swapDelete(i int) {
	last := len(*s) - 1
	(*s)[i] = (*s)[last]
	*s = (*s)[:last]
}

func (s *storeImpl[C]) append(data C) int {
	*s = append(*s, data)
	return len(*s) - 1
}
