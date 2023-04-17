package reflect

import (
	"reflect"

	"github.com/Tnze/go-ecs"
)

func NumComps(w *ecs.World, e ecs.Entity) int {
	return len(w.Entities[e].AT.Types)
}

func IndexComps(w *ecs.World, e ecs.Entity, i int) (c ecs.Component, v reflect.Value) {
	rec := w.Entities[e]
	c = rec.AT.Types[i].Component
	if store := reflect.ValueOf(rec.AT.Comps[i]); !store.IsNil() {
		v = store.Index(rec.Row)
	}
	return c, v
}
