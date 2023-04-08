package ecsutil

import (
	"reflect"

	"github.com/Tnze/go-ecs"
)

type World struct {
	*ecs.World
	components map[reflect.Type]ecs.Component
}

func NewWorld() *World {
	return &World{
		World:      ecs.NewWorld(),
		components: make(map[reflect.Type]ecs.Component),
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
		c = ecs.NewNamedComponent(e.w.World, t.String())
	}
	ecs.Set(e.w.World, e.Entity, c, data)
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
	ecs.Remove(e.w.World, e.Entity, c)
}
