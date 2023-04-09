package warp

import (
	"reflect"

	"github.com/Tnze/go-ecs"
)

type World struct {
	*ecs.World
	components map[reflect.Type]ecs.Component
	NameComp   ecs.Component
}

func NewWorld() *World {
	w := ecs.NewWorld()
	name := ecs.NewComponent(w)
	ecs.SetComp(w, name.Entity, name, "Name")
	return &World{
		World:      w,
		components: make(map[reflect.Type]ecs.Component),
		NameComp:   name,
	}
}

func (w *World) NewEntity() Entity {
	return Entity{ecs.NewEntity(w.World), w}
}

func (w *World) NewNamedEntity(name string) Entity {
	e := ecs.NewEntity(w.World)
	ecs.SetComp(w.World, e, w.NameComp, name)
	return Entity{e, w}
}

func (w *World) NewComponent() Component {
	return Component{
		Component: ecs.NewComponent(w.World),
		w:         w,
	}
}

func (w *World) NewNamedComponent(name string) Component {
	c := ecs.NewComponent(w.World)
	ecs.SetComp(w.World, c.Entity, w.NameComp, name)
	return Component{
		Component: c,
		w:         w,
	}
}

type Entity struct {
	ecs.Entity
	w *World
}

func (e *Entity) Name() *string {
	return ecs.Get[string](e.w.World, e.Entity, e.w.NameComp)
}

func (e *Entity) Del() {
	ecs.DelEntity(e.w.World, e.Entity)
	e.Entity = 0
	e.w = nil
}

func SetComp[C any](e Entity, data C) {
	t := reflect.TypeOf(data)
	c, ok := e.w.components[t]
	if !ok {
		// c = e.w.NewNamedComponent(t.String()).Component
		c = e.w.NewComponent().Component
		e.w.components[t] = c
	}
	ecs.SetComp(e.w.World, e.Entity, c, data)
}

func GetComp[C any](e Entity) (data *C) {
	c, ok := e.w.components[reflect.TypeOf(data).Elem()]
	if !ok {
		return nil
	}
	return ecs.Get[C](e.w.World, e.Entity, c)
}

func AddComp[C any](e Entity) {
	var tmpC C
	t := reflect.TypeOf(tmpC)
	c, ok := e.w.components[t]
	if !ok {
		// c = e.w.NewNamedComponent(t.String()).Component
		c = e.w.NewComponent().Component
		e.w.components[t] = c
	}
	ecs.AddComp(e.w.World, e.Entity, c)
}

func DelComp[C any](e Entity) {
	var tmpC C
	c, ok := e.w.components[reflect.TypeOf(tmpC)]
	if !ok {
		return
	}
	ecs.DelComp(e.w.World, e.Entity, c)
}

type Component struct {
	ecs.Component
	w *World
}
