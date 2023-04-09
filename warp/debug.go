package warp

import "github.com/Tnze/go-ecs"

func (e Entity) Type() string {
	return ecs.Type(e.w.World, e.Entity, e.w.NameComp)
}
