package ecs

type Filter func(*World, *archetype, *[]int) bool

func (f Filter) Run(w *World, h func(entities Table[Entity], data []any)) {
	var columns []int
	var data []any
	for _, a := range w.archetypes {
		columns = columns[:0]
		if !f(w, a, &columns) {
			continue
		}
		data = data[:0]
		for _, col := range columns {
			data = append(data, a.comps[col])
		}
		h(a.entities, data)
	}
}

func QueryAll(comps ...Component) Filter {
	return func(w *World, a *archetype, out *[]int) bool {
		for _, c := range comps {
			col, ok := w.components[c][a]
			if !ok {
				return false
			}
			// Empty components are excluded from the output.
			if a.comps[col] != nil {
				*out = append(*out, col)
			}
		}
		return true
	}
}

func QueryAny(comps ...Component) Filter {
	return func(w *World, a *archetype, out *[]int) (pass bool) {
		for _, c := range comps {
			if col, ok := w.components[c][a]; ok {
				// Empty components are excluded from the output.
				if a.comps[col] != nil {
					*out = append(*out, col)
				}
				pass = true
			}
		}
		return
	}
}

func (f Filter) Cache(w *World) (q *CachedQuery) {
	var columns [][]int
	var tables []*archetype

	var out []int
	for _, a := range w.archetypes {
		out = make([]int, 0, len(out))
		if f(w, a, &out) {
			columns = append(columns, out)
			tables = append(tables, a)
		}
	}

	q = &CachedQuery{
		filter:  f,
		tables:  tables,
		columns: columns,
		data:    nil,
	}
	q.row = w.queries.append(q)
	return q
}

// CachedQuery is cached filter
type CachedQuery struct {
	filter  Filter
	tables  []*archetype // All archetypes in the world that matches the filter.
	columns [][]int      // For each archetype, the storage indexes for every component data.

	// Cached arguments for the callback, to avoid allocating memory every time Run is called.
	data []any
	row  int // self index in World.queries
}

func (q *CachedQuery) Run(h func(entities Table[Entity], data []any)) {
	data := q.data[:0]
	for i, a := range q.tables {
		data = data[:0]
		for _, col := range q.columns[i] {
			data = append(data, a.comps[col])
		}
		h(a.entities, data)
	}
	q.data = data
}

func (q *CachedQuery) update(w *World, a *archetype) {
	var numOfCol int
	if len(q.columns) > 0 {
		numOfCol = len(q.columns[0])
	}

	out := make([]int, 0, numOfCol)
	if q.filter(w, a, &out) {
		q.columns = append(q.columns, out)
		q.tables = append(q.tables, a)
	}
}

func (q *CachedQuery) Free(w *World) {
	w.queries.swapDelete(q.row)
	if q.row < len(w.queries) {
		w.queries[q.row].row = q.row
	}
}
