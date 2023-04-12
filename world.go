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

		// The default archetype for newly created entities, which contains no components.
		zero *archetype

		// All entities in the World, including components.
		// Records their archetype's pointer and the index of the comps belonging to the entity.
		entities map[Entity]*entityRecord

		// All archetypes in the World.
		// The key of the map is the hash of the archetype's types.
		// And the value is the archetype's pointer.
		archetypes map[uint64]*archetype

		// This field stores maps for each component.
		// Each map contains a list of archetypes that have the component.
		// And the component's corresponding column index in the archetype.
		//
		// We can check if an archetype has a component by looking up the map.
		//
		// For any Component c and archetype a:
		//	col, ok := components[c][a]
		// If ok == true, then archetype a has component c, otherwise it doesn't.
		// And if col == -1, archetype a has component c but doesn't contain any data,
		// otherwise the col is the index of the component's column in the archetype.
		components map[Component]map[*archetype]int
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
		nextID   uint64
		freelist []uint64
	}
	entityRecord struct {
		at  *archetype
		row int
	}
	archetype struct {
		// A list of components.
		// It's sorted and able to be hashed.
		// Allowing us to find the archetype by the hash of its type.
		types
		entities Table[Entity]
		records  Table[*entityRecord]
		comps    []column

		// A list of edges to other archetypes.
		// Used to find the next archetype when adding or removing components.
		edges map[Component]archetypeEdge
	}
	types         []componentMeta
	componentMeta struct {
		Component
		// If the component's corresponding data has type T,
		// this stores the reflect.Type of Table[T].
		// We need this because, when creating new archetypes,
		// we need to create new storage for the components.
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
	Table[C any] []C
)

// NewWorld creates a new empty World, with the default components.
func NewWorld() (w *World) {
	w = &World{
		entities:   make(map[Entity]*entityRecord),
		archetypes: make(map[uint64]*archetype),
		components: make(map[Component]map[*archetype]int),
	}
	w.zero = newArchetype(w, nil)
	return
}

// NewEntity creates a new Entity in the World, without any components.
func NewEntity(w *World) (e Entity) {
	e = Entity(w.get())
	r := new(entityRecord)
	r.at = w.zero
	r.row = w.zero.entities.append(e)
	w.zero.records.append(r)
	w.entities[e] = r
	return
}

func DelEntity(w *World, e Entity) {
	rec := w.entities[e]
	rec.at.entities.swapDelete(rec.row)
	rec.at.records.swapDelete(rec.row)
	for _, t := range rec.at.types {
		if col := w.components[t.Component][rec.at]; col >= 0 {
			rec.at.comps[col].swapDelete(rec.row)
		}
	}
	if rec.row != len(rec.at.entities) {
		rec.at.records[rec.row].row = rec.row
	}
	delete(w.entities, e)
	w.idManager.put(uint64(e))
}

// NewComponent creates a new Component in the World.
// The data type associated with the Component will be bind when the first data is set.
func NewComponent(w *World) (c Component) {
	c = Component{NewEntity(w)}
	w.components[c] = make(map[*archetype]int)
	return
}

func newArchetype(w *World, t types) (a *archetype) {
	a = &archetype{
		types: t,
		comps: make([]column, len(t)),
		edges: make(map[Component]archetypeEdge),
	}
	for i, v := range t {
		if v.columnType == nil {
			w.components[v.Component][a] = -1
		} else {
			a.comps[i] = reflect.New(v.columnType.Elem()).Interface().(column)
			w.components[v.Component][a] = i
		}
	}
	w.archetypes[t.sortHash()] = a
	return
}

func moveEntity(w *World, dst *archetype, srcRec *entityRecord, list types) (newRow int) {
	e := srcRec.at.entities[srcRec.row]
	newRow = dst.entities.append(e)
	dst.records.append(srcRec)
	srcRec.at.entities.swapDelete(srcRec.row)
	srcRec.at.records.swapDelete(srcRec.row)
	// Move other components
	for _, t := range list {
		idx := w.components[t.Component]
		if col := idx[srcRec.at]; col >= 0 {
			src := srcRec.at.comps[col]
			dst.comps[idx[dst]].appendFrom(src, srcRec.row)
			src.swapDelete(srcRec.row)
		}
	}
	return
}

// AddComp adds the Component to Entity as a label, with no data.
func AddComp(w *World, e Entity, c Component) {
	rec := w.entities[e]
	// If the archetype of e already contains c, return.
	if _, ok := w.components[c][rec.at]; ok {
		return
	}
	// Lookup archetypeEdge for shortcuts
	edge := rec.at.edges[c]
	target := edge.add
	if target == nil {
		// We don't have shortcuts yet. Use the hash way.
		var ok bool
		newTypes := rec.at.types.copyAppend(c, nil)
		if target, ok = w.archetypes[newTypes.sortHash()]; !ok {
			target = newArchetype(w, newTypes)
		}
		// Save to the shortcuts
		edge.add = target
		rec.at.edges[c] = edge
	}
	// Move entity to the new archetype
	row := moveEntity(w, target, rec, rec.at.types)
	// Because we move the last entity in rec.at.entities.
	// We have to update its row value in w.entities.
	if rec.row != len(rec.at.entities) {
		rec.at.records[rec.row].row = rec.row
	}
	if col := w.components[c][target]; col >= 0 {
		panic("cannot add component has type" + target.types[col].columnType.String())
	}

	rec.at = target
	rec.row = row
}

func HasComp(w *World, e Entity, c Component) bool {
	rec := w.entities[e]
	_, ok := w.components[c][rec.at]
	return ok
}

// SetComp sets the data of a Component of an Entity.
//
// If the Entity already has the Component, the data will be overridden.
// If the Entity doesn't have the Component, the Component will be added.
//
// This function panics if the type of data doesn't match others of the same Component.
func SetComp[C any](w *World, e Entity, c Component, data C) {
	rec := w.entities[e]
	// If the archetype of e already contains c.
	// Override the data and return.
	if col, ok := w.components[c][rec.at]; ok {
		(*rec.at.comps[col].(*Table[C]))[rec.row] = data
		return
	}
	// Lookup archetypeEdge for shortcuts
	edge := rec.at.edges[c]
	target := edge.add
	if target == nil {
		// We don't have shortcuts yet. Use the hash way.
		var tmpS *Table[C]
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
	row := moveEntity(w, target, rec, rec.at.types)
	// Because we move the last entity in rec.at.entities.
	// We have to update its row value in w.entities.
	if rec.row != len(rec.at.entities) {
		rec.at.records[rec.row].row = rec.row
	}
	target.comps[w.components[c][target]].(*Table[C]).append(data)

	rec.at = target
	rec.row = row
}

// DelComp removes the Component of an Entity.
// If the Entity doesn't have the Component, nothing will happen.
func DelComp(w *World, e Entity, c Component) {
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
	row := moveEntity(w, target, rec, target.types)
	// Because we move the last entity in rec.at.entities.
	// We have to update its row value in w.entities.
	if rec.row != len(rec.at.entities) {
		rec.at.records[rec.row].row = rec.row
	}

	rec.at = target
	rec.row = row
}

// Get gets the data of a Component of an Entity.
// If the Entity doesn't have the Component, nil will be returned.
func Get[C any](w *World, e Entity, c Component) (data *C) {
	rec := w.entities[e]
	if column, ok := w.components[c][rec.at]; ok {
		return &(*rec.at.comps[column].(*Table[C]))[rec.row]
	}
	return nil
}

func (i *idManager) get() (id uint64) {
	if length := len(i.freelist); length > 0 {
		id = i.freelist[length-1]
		i.freelist = i.freelist[:length-1]
		return
	}
	id = i.nextID
	i.nextID++
	return
}

func (i *idManager) put(id uint64) {
	i.freelist = append(i.freelist, id)
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

func (c Table[C]) len() int {
	return len(c)
}

func (c *Table[C]) appendFrom(other column, row int) {
	*c = append(*c, (*other.(*Table[C]))[row])
}

func (c *Table[C]) swapDelete(i int) {
	last := len(*c) - 1
	(*c)[i] = (*c)[last]
	*c = (*c)[:last]
}

func (c *Table[C]) append(data C) int {
	*c = append(*c, data)
	return len(*c) - 1
}
