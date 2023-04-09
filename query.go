package ecs

func TermIter[C any](w *World, c Component, f func(entities []Entity, fields []C)) {
	for a, column := range w.components[c] {
		f(a.entities, *a.comps[column].(*Table[C]))
	}
}

type Filter []Component

func (f Filter) All(w *World, h func(entities []Entity, data []any)) {
	columns := make([]int, len(f))
	data := make([]any, 0, len(f))
a:
	for _, a := range w.archetypes {
		for i, c := range f {
			var ok bool
			if columns[i], ok = w.components[c][a]; !ok {
				continue a
			}
		}
		// all matched! good!
		for _, col := range columns {
			if col >= 0 {
				data = append(data, a.comps[col])
			}
		}
		h(a.entities, data)
		data = data[:0]
	}
}

func (f Filter) Any() {
}

func (f Filter) Nor() {}
