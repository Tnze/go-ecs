package ecs

func TermIter[C any](w *World, c Component, f func(field []C)) {
	for a, column := range w.components[c] {
		f(*a.comps[column].(*storeImpl[C]))
	}
}
