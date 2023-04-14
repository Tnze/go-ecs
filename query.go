package ecs

type Filter func(*World, *archetype, *[]int) bool

func (f Filter) Cache(w *World) *CachedQuery {
	q := &CachedQuery{filter: f}
	var columns []int
	for _, a := range w.archetypes {
		columns = columns[:0]
		if q.filter(w, a, &columns) {
			columns2 := make([]int, len(columns))
			copy(columns2, columns)
			q.columns = append(q.columns, columns2)
			q.tables = append(q.tables, a)
		}
	}
	return q
}

func (f Filter) Run(w *World, h func(entities Table[Entity], data []any)) {
	var columns []int
	var data []any
	for _, a := range w.archetypes {
		columns = columns[:0]
		data = data[:0]
		if f(w, a, &columns) {
			for _, col := range columns {
				data = append(data, a.comps[col])
			}
			h(a.entities, data)
		}
	}
}

func QueryAll(comps ...Component) Filter {
	return func(w *World, a *archetype, out *[]int) bool {
		for _, c := range comps {
			col, ok := w.components[c][a]
			if !ok {
				return false
			}
			*out = append(*out, col)
		}
		return true
	}
}

func QueryAny(comps ...Component) Filter {
	return func(w *World, a *archetype, out *[]int) (pass bool) {
		for _, c := range comps {
			if col, ok := w.components[c][a]; ok {
				pass = true
				*out = append(*out, col)
			}
		}
		return
	}
}

// CachedQuery is cached filter
//
// BUG(Tnze): Currently the query doesn't get updated when a new archetype is created.
type CachedQuery struct {
	filter  Filter
	tables  []*archetype
	columns [][]int

	data []any
}

func (q *CachedQuery) Run(h func(entities Table[Entity], data []any)) {
	data := q.data[:0]
	for i, a := range q.tables {
		for _, col := range q.columns[i] {
			data = append(data, a.comps[col])
		}
		h(a.entities, data)
	}
	q.data = data
}
