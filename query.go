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

func (f Filter) Any() {
}

func (f Filter) Nor() {}

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
	return &q
}

func (f Filter) CacheAny(w *World) *Query {
	return nil
}

func (f Filter) CacheNor(w *World) *Query {
	return nil
}

// Query is cached filter
//
// BUG(Tnze): Currently the query doesn't get updated when a new archetype is created.
type Query struct {
	world   *World
	filter  Filter
	tables  []*archetype
	columns [][]int
}

func (q *Query) Run(h func(entities Table[Entity], data []any)) {
	data := make([]any, 0, len(q.filter))
	for i, a := range q.tables {
		for j := range q.filter {
			if col := q.columns[i][j]; col >= 0 {
				data = append(data, a.comps[col])
			}
		}
		h(a.entities, data)
	}
}
