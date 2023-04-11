package ecs

type Filter []Component

func (f Filter) All(w *World, h func(entities Table[Entity], data []any)) {
	data := make([]any, 0, len(f))
a:
	for _, a := range w.archetypes {
		data = data[:0]
		for _, c := range f {
			col, ok := w.components[c][a]
			if !ok {
				continue a
			}
			if col >= 0 {
				data = append(data, a.comps[col])
			}
		}
		h(a.entities, data)
	}
}

func (f Filter) Any(w *World, h func(entities Table[Entity], data []any)) {
	data := make([]any, 0, len(f))
	for _, a := range w.archetypes {
		var pass bool
		data = data[:0]
		for _, c := range f {
			if col, ok := w.components[c][a]; ok {
				pass = true
				if col >= 0 {
					data = append(data, a.comps[col])
				}
			}
		}
		if pass {
			h(a.entities, data)
		}
	}
}

func (f Filter) CacheAll(w *World) *Query {
	var q Query
a:
	for _, a := range w.archetypes {
		columns := make([]int, len(f))
		for i, c := range f {
			col, ok := w.components[c][a]
			if !ok {
				continue a
			}
			columns[i] = col
		}
		q.columns = append(q.columns, columns)
		q.tables = append(q.tables, a)
	}
	q.world = w
	q.filter = f
	q.data = make([]any, 0, len(f))
	return &q
}

// Query is cached filter
//
// BUG(Tnze): Currently the query doesn't get updated when a new archetype is created.
type Query struct {
	world   *World
	filter  Filter
	tables  []*archetype
	columns [][]int

	data []any
}

func (q *Query) Run(h func(entities Table[Entity], data []any)) {
	data := q.data[:0]
	for i, a := range q.tables {
		for _, col := range q.columns[i] {
			data = append(data, a.comps[col])
		}
		h(a.entities, data)
	}
	q.data = data
}
