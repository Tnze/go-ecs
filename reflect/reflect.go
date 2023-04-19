package reflect

import (
	"github.com/Tnze/go-ecs"
	"github.com/Tnze/go-ecs/internal/core"
)

type Value struct {
	rec *core.EntityRecord
}

func ValueOf(w *ecs.World, e ecs.Entity) Value {
	return Value{w.Entities[e]}
}

func (v Value) NumComps() int {
	return len(v.rec.AT.Types)
}

func (v Value) IndexComps(i int) (c ecs.Component, val any) {
	c = v.rec.AT.Types[i].Component
	if store := v.rec.AT.Comps[i]; store != nil {
		val = store.Get(v.rec.Row)
	}
	return c, val
}
