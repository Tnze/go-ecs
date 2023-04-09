package ecsutil

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

type Entity struct {
	ecs.Entity
	w *World
}

func Set[C any](e Entity, data C) {
	t := reflect.TypeOf(data)
	c, ok := e.w.components[t]
	if !ok {
		c = ecs.NewComponent(e.w.World)
		ecs.SetComp(e.w.World, e.Entity, e.w.NameComp, t.String())
	}
	ecs.SetComp(e.w.World, e.Entity, c, data)
}

func Get[C any](e Entity) (data *C) {
	c, ok := e.w.components[reflect.TypeOf(data).Elem()]
	if !ok {
		return nil
	}
	return ecs.Get[C](e.w.World, e.Entity, c)
}

func Remove[C any](e Entity) {
	var tmpC C
	c, ok := e.w.components[reflect.TypeOf(tmpC)]
	if !ok {
		return
	}
	ecs.DelComp(e.w.World, e.Entity, c)
}
