package ecs

import "github.com/Tnze/go-ecs/internal/core"

type (
	World     = core.World
	Entity    = core.Entity
	Component = core.Component

	Filter      func(*World, *core.Archetype, *[]int) bool
	CachedQuery core.CachedQuery
)

func NewWorld() (w *World)            { return core.NewWorld() }
func NewEntity(w *World) Entity       { return core.NewEntity(w) }
func DelEntity(w *World, e Entity)    { core.DelEntity(w, e) }
func NewComponent(w *World) Component { return core.NewComponent(w) }

func AddComp(w *World, e Entity, c Component)                  { core.AddComp(w, e, c) }
func HasComp(w *World, e Entity, c Component) bool             { return core.HasComp(w, e, c) }
func SetComp[C any](w *World, e Entity, c Component, data C)   { core.SetComp[C](w, e, c, data) }
func GetComp[C any](w *World, e Entity, c Component) (data *C) { return core.GetComp[C](w, e, c) }
func DelComp(w *World, e Entity, c Component)                  { core.DelComp(w, e, c) }

func QueryAll(comps ...Component) Filter { return Filter(core.QueryAll(comps...)) }
func QueryAny(comps ...Component) Filter { return Filter(core.QueryAny(comps...)) }

func (f Filter) Run(w *World, h func([]Entity, []any)) { core.Filter(f).Run(w, h) }
func (f Filter) Cache(w *World) *CachedQuery           { return (*CachedQuery)(core.Filter(f).Cache(w)) }
func (q *CachedQuery) Run(h func([]Entity, []any))     { (*core.CachedQuery)(q).Run(h) }
func (q *CachedQuery) Free(w *World)                   { (*core.CachedQuery)(q).Free(w) }

// debug

func Type(w *World, e Entity, nameComp Component) string { return core.Type(w, e, nameComp) }
