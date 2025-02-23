package ecs

import (
	"hash/maphash"
	"reflect"
	"sort"
	"unsafe"
)

type (
	// The World is the container for all ECS data.
	// It stores the entities and their Components, does queries and runs systems.
	//
	// --flecs.dev
	World struct {
		IDManager
		hash maphash.Hash // used for hash archetypes' type

		// The default archetype for newly created entities, which contains no Components.
		Zero *Archetype

		// All entities in the World, including Components.
		// Records their archetype's pointer and the index of the Comps belonging to the entity.
		Entities map[Entity]*EntityRecord

		// All archetypes in the World.
		// The key of the map is the hash of the archetype's Types.
		// And the value is the archetype's pointer.
		Archetypes map[uint64]*Archetype

		// This field stores maps for each component.
		// Each map contains a list of archetypes that have the component.
		// And the component's corresponding Storage index in the archetype.
		//
		// We can check if an archetype has a component by looking up the map.
		//
		// For any Component c and archetype a:
		//	col, ok := Components[c][a]
		// If ok == true, then archetype a has component c, otherwise it doesn't.
		// And if col == -1, archetype a has component c but doesn't contain any data,
		// otherwise the col is the index of the component's Storage in the archetype.
		Components map[Component]map[*Archetype]int

		// For high performance, we cache the queries.
		// But these caches will get outdated when new archetypes are created.
		// We register all queries created here, and update them when new archetypes are created.
		Queries Table[*CachedQuery]
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
	Component Entity

	// The IDManager is an internal structure which is used to generate/recycle entity IDs.
	IDManager struct {
		NextID   uint64
		Freelist []uint64
	}
	EntityRecord struct {
		AT  *Archetype
		Row int
	}
	Archetype struct {
		Types
		entities Table[Entity]
		records  Table[*EntityRecord]
		Comps    []Storage

		// A list of edges to other archetypes.
		// Used to find the next archetype when adding or removing Components.
		edges map[Component]ArchetypeEdge
	}

	// Types is list of Components.
	// It's sorted and able to be hashed.
	// Allowing us to find the archetype by the hash of its type.
	Types         []ComponentMeta
	ComponentMeta struct {
		Component
		// This stores the reflect.Type of *Table[T],
		// which T is the type of component's corresponding data.
		// We need this because, when creating new archetypes,
		// we need to create new Storage for the Components.
		TableType reflect.Type
	}
	ArchetypeEdge struct {
		add, del *Archetype
	}
	Storage interface {
		appendFrom(other Storage, column int)
		swapDelete(i int)
		toSlice() any

		Get(i int) any
	}
	Table[C any] []C
)

// NewWorld creates a new empty World, with the default Components.
func NewWorld() (w *World) {
	w = &World{
		Entities:   make(map[Entity]*EntityRecord),
		Archetypes: make(map[uint64]*Archetype),
		Components: make(map[Component]map[*Archetype]int),
	}
	w.Zero = newArchetype(w, Types(nil), Types(nil).sortHash(&w.hash))
	return
}

// NewEntity creates a new Entity in the World, without any Components.
func NewEntity(w *World) (e Entity) {
	e = Entity(w.get())
	r := new(EntityRecord)
	r.AT = w.Zero
	r.Row = w.Zero.entities.append(e)
	w.Zero.records.append(r)
	w.Entities[e] = r
	return
}

func DelEntity(w *World, e Entity) {
	rec := w.Entities[e]
	rec.AT.entities.swapDelete(rec.Row)
	rec.AT.records.swapDelete(rec.Row)
	for _, s := range rec.AT.Comps {
		if s != nil {
			s.swapDelete(rec.Row)
		}
	}
	if rec.Row != len(rec.AT.entities) {
		rec.AT.records[rec.Row].Row = rec.Row
	}
	delete(w.Entities, e)
	w.IDManager.put(uint64(e))
}

// NewComponent creates a new Component in the World.
// The data type associated with the Component will be bind when the first data is set.
func NewComponent(w *World) (c Component) {
	c = Component(NewEntity(w))
	w.Components[c] = make(map[*Archetype]int)
	return
}

// The hash can be calculated by t.sortHash(&w.hash).
// We Always calculate it before calling this function, so just pass it in.
func newArchetype(w *World, t Types, hash uint64) (a *Archetype) {
	a = &Archetype{
		Types: t,
		Comps: make([]Storage, len(t)),
		edges: make(map[Component]ArchetypeEdge),
	}
	for i, v := range t {
		if v.TableType != nil {
			a.Comps[i] = reflect.New(v.TableType.Elem()).Interface().(Storage)
			w.Components[v.Component][a] = i
		} else {
			w.Components[v.Component][a] = -1
		}
	}
	w.Archetypes[hash] = a

	// update queries
	for _, q := range w.Queries {
		q.update(w, a)
	}
	return
}

// AddComp adds the Component to Entity as a tag, without underlying content
func AddComp(w *World, e Entity, c Component) {
	rec := w.Entities[e]
	// If the archetype of e already contains c.
	// Override the data and return.
	if _, ok := w.Components[c][rec.AT]; ok {
		return
	}
	// Lookup ArchetypeEdge for shortcuts
	edge := rec.AT.edges[c]
	target := edge.add
	if target == nil {
		// We don't have shortcuts yet. Use the hash way.
		var ok bool
		newTypes := rec.AT.Types.copyAppend(c, nil)
		hash := newTypes.sortHash(&w.hash)
		if target, ok = w.Archetypes[hash]; !ok {
			target = newArchetype(w, newTypes, hash)
		}
		// Save to the shortcuts
		edge.add = target
		rec.AT.edges[c] = edge
	}
	// Move entity to the new archetype
	row := moveEntity(e, target, rec, rec.AT.Types)
	// Because we move the last entity in rec.AT.entities.
	// We have to update its Row value in w.entities.
	if rec.Row != len(rec.AT.entities) {
		rec.AT.records[rec.Row].Row = rec.Row
	}

	rec.AT = target
	rec.Row = row
}

// HasComp reports whether the Entity has the Component.
func HasComp(w *World, e Entity, c Component) bool {
	rec := w.Entities[e]
	_, ok := w.Components[c][rec.AT]
	return ok
}

// SetComp adds the Component and its content to Entity.
//
// If the Entity already has the Component, the content will be overridden.
// If the Entity doesn't have the Component, the Component will be added.
//
// This function panics if the type of data doesn't match others of the same Component.
func SetComp[C any](w *World, e Entity, c Component, data C) {
	rec := w.Entities[e]
	// If the archetype of e already contains c.
	// Override the data and return.
	if col, ok := w.Components[c][rec.AT]; ok {
		(*rec.AT.Comps[col].(*Table[C]))[rec.Row] = data
		return
	}
	// Lookup ArchetypeEdge for shortcuts
	edge := rec.AT.edges[c]
	target := edge.add
	if target == nil {
		// We don't have shortcuts yet. Use the hash way.
		var ok bool
		newTypes := rec.AT.Types.copyAppend(c, reflect.TypeFor[*Table[C]]())
		hash := newTypes.sortHash(&w.hash)
		if target, ok = w.Archetypes[hash]; !ok {
			target = newArchetype(w, newTypes, hash)
		}
		// Save to the shortcuts
		edge.add = target
		rec.AT.edges[c] = edge
	}
	// Move entity to the new archetype
	row := moveEntity(e, target, rec, rec.AT.Types)
	// Because we move the last entity in rec.AT.entities.
	// We have to update its Row value in w.entities.
	if rec.Row != len(rec.AT.entities) {
		rec.AT.records[rec.Row].Row = rec.Row
	}
	target.Comps[w.Components[c][target]].(*Table[C]).append(data)

	rec.AT = target
	rec.Row = row
}

// DelComp removes the Component of an Entity.
// If the Entity doesn't have the Component, nothing will happen.
func DelComp(w *World, e Entity, c Component) {
	rec := w.Entities[e]
	col, ok := w.Components[c][rec.AT]
	if !ok {
		return // archetype of e doesn't contain component c
	}
	// Lookup ArchetypeEdge for shortcuts
	edge := rec.AT.edges[c]
	target := edge.del
	if target == nil {
		// We don't have shortcuts yet. Use the hash way.
		newTypes := rec.AT.Types.copyDelete(col)
		hash := newTypes.sortHash(&w.hash)
		if target, ok = w.Archetypes[hash]; !ok {
			target = newArchetype(w, newTypes, hash)
		}
		// Save to the shortcuts
		edge.del = target
		rec.AT.edges[c] = edge
	}
	// Move entity
	row := moveEntity(e, target, rec, target.Types)
	// Because we move the last entity in rec.AT.entities.
	// We have to update its Row value in w.entities.
	if rec.Row != len(rec.AT.entities) {
		rec.AT.records[rec.Row].Row = rec.Row
	}

	rec.AT = target
	rec.Row = row
}

func moveEntity(e Entity, dst *Archetype, srcRec *EntityRecord, list Types) (newRow int) {
	// Copy Components
	srcCol, dstCol := 0, 0
	for _, t := range list {
		for srcRec.AT.Types[srcCol].Component != t.Component {
			// The Types are ordered, and srcRec.AT must contain all Components in the list.
			// So we surely will find the correct index, and the access to Types[i] won't panic.
			srcCol++
		}
		for dst.Types[dstCol].Component != t.Component {
			// So do here.
			dstCol++
		}
		if src := srcRec.AT.Comps[srcCol]; src != nil {
			dst.Comps[dstCol].appendFrom(src, srcRec.Row)
		}
	}
	// Delete everything in src
	newRow = dst.entities.append(e)
	dst.records.append(srcRec)
	srcRec.AT.entities.swapDelete(srcRec.Row)
	srcRec.AT.records.swapDelete(srcRec.Row)
	for _, s := range srcRec.AT.Comps {
		if s != nil {
			s.swapDelete(srcRec.Row)
		}
	}
	return
}

// GetComp gets the data of a Component of an Entity.
// If the Entity doesn't have the Component, nil will be returned.
func GetComp[C any](w *World, e Entity, c Component) (data *C) {
	rec := w.Entities[e]
	if column, ok := w.Components[c][rec.AT]; ok {
		return &(*rec.AT.Comps[column].(*Table[C]))[rec.Row]
	}
	return nil
}

// get an ID from the IDManager.
// If the Freelist isn't empty, the ID is obtained there, otherwise it's generated incrementally.
func (i *IDManager) get() (id uint64) {
	if length := len(i.Freelist); length > 0 {
		id = i.Freelist[length-1]
		i.Freelist = i.Freelist[:length-1]
	} else {
		id = i.NextID
		i.NextID++
	}
	return
}

// put an ID into the IDManager.
// The ID will be recycled and stored in the Freelist, and to be reused later.
func (i *IDManager) put(id uint64) {
	i.Freelist = append(i.Freelist, id)
}

func (t Types) sortHash(hash *maphash.Hash) uint64 {
	// sort the component list
	if t != nil {
		sort.Slice(t, func(i, j int) bool {
			return t[i].Component < t[j].Component
		})
	}
	// calculate hash
	hash.Reset()
	for i := range t {
		hash.Write((*[8]byte)(unsafe.Pointer(&t[i].Component))[:])
	}
	return hash.Sum64()
}

func (t Types) copyAppend(c Component, storeType reflect.Type) (newTypes Types) {
	newTypes = make(Types, len(t)+1)
	newTypes[0] = ComponentMeta{
		Component: c,
		TableType: storeType,
	}
	copy(newTypes[1:], t)
	return
}

func (t Types) copyDelete(i int) (newTypes Types) {
	newTypes = make(Types, len(t)-1)
	copy(newTypes[:i], t[:i])
	if i+1 < len(t) {
		copy(newTypes[i:], t[i+1:])
	}
	return
}

func (c *Table[C]) appendFrom(other Storage, row int) {
	*c = append(*c, (*other.(*Table[C]))[row])
}

func (c *Table[C]) swapDelete(i int) {
	last := len(*c) - 1
	(*c)[i] = (*c)[last]
	*c = (*c)[:last]
}

func (c *Table[C]) toSlice() any {
	return (*[]C)(c)
}

func (c *Table[C]) Get(i int) any {
	return (*c)[i]
}

func (c *Table[C]) append(data C) int {
	*c = append(*c, data)
	return len(*c) - 1
}
