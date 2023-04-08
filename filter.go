package ecs

func TermIter[C any](w *World, c Component, f func(entities []Entity, fields []C)) {
	for a, column := range w.components[c] {
		f(a.entities, *a.comps[column].(*columnImpl[C]))
	}
}
