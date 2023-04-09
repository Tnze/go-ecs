package warp

import (
	"fmt"
	"reflect"

	"github.com/Tnze/go-ecs"
)

func (w *World) Filter(h any) {
	rh := reflect.ValueOf(h)
	rt := rh.Type()
	numIn := rt.NumIn()
	if numIn < 1 {
		panic("a filter handler must receive a []Entity as its first argument")
	}
	filter := make(ecs.Filter, numIn-1)
	var ok bool
	for i := range filter {
		ts := rt.In(i + 1)
		if ts.Kind() != reflect.Slice {
			panic(fmt.Errorf("the %d(st/nd/rd/th) argument has type %v, not a slice", i+1, ts))
		}
		tc := ts.Elem()
		filter[i], ok = w.components[tc]
		if !ok {
			panic("cannot find Component corresponding to type " + tc.String())
		}
	}
	filter.All(w.World, func(entities ecs.Table[ecs.Entity], data []any) {
		in := make([]reflect.Value, 1+len(data))
		warpEntities := make([]Entity, len(entities))
		in[0] = reflect.ValueOf(warpEntities)
		for i := range entities {
			warpEntities[i] = Entity{Entity: entities[i], w: w}
		}
		for i, v := range data {
			in[i+1] = reflect.ValueOf(v).Elem()
		}
		rh.Call(in)
	})
}
