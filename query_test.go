package ecs

import (
	"math/rand"
	"reflect"
	"sort"
	"testing"
)

func TestFilter_Run(t *testing.T) {
	w := NewWorld()

	// Create 10 entities.
	var entities [10]Entity
	for i := range entities {
		entities[i] = NewEntity(w)
	}

	// Create 2 components.
	c1 := NewComponent(w)
	c2 := NewComponent(w)

	// Add components to entities.
	for i, e := range entities[:5] {
		SetComp(w, e, c1, i)
	}
	for i, e := range entities[3:7] {
		SetComp(w, e, c2, i+3)
	}

	// id: [0 1 2 3 4 5 6 7 8 9]
	// c1: [0 1 2 3 4          ]
	// c2: [      3 4 5 6      ]

	filters := [][]Component{{c1}, {c2}, {c1, c2}}
	var wants [][]int
	testAll := func(t *testing.T) {
		var result []int
		for i, want := range wants {
			result = result[:0]
			QueryAll(filters[i]...).Run(w, func(entities Table[Entity], data []any) {
				result = append(result, []int(*data[0].(*Table[int]))...)
			})
			// The order of the results is not guaranteed, so sort them before validation
			sort.Ints(result)
			if !reflect.DeepEqual(result, want) {
				t.Errorf("get: %v, want: %v", result, want)
			}
		}
	}
	testAny := func(t *testing.T) {
		var result []int
		for i, want := range wants {
			result = result[:0]
			QueryAny(filters[i]...).Run(w, func(entities Table[Entity], data []any) {
				result = append(result, []int(*data[0].(*Table[int]))...)
			})
			// The order of the results is not guaranteed, so sort them before validation
			sort.Ints(result)
			if !reflect.DeepEqual(result, want) {
				t.Errorf("get: %v, want: %v", result, want)
			}
		}
	}

	wants = [][]int{{0, 1, 2, 3, 4}, {3, 4, 5, 6}, {3, 4}}
	t.Run("All", testAll)
	wants = [][]int{{0, 1, 2, 3, 4}, {3, 4, 5, 6}, {0, 1, 2, 3, 4, 5, 6}}
	t.Run("Any", testAny)

	// change the entities
	SetComp(w, entities[6], c1, 6)
	DelComp(w, entities[3], c2)

	// id: [0 1 2 3 4 5 6 7 8 9]
	// c1: [0 1 2 3 4   6      ]
	// c2: [      _ 4 5 6      ]

	wants = [][]int{{0, 1, 2, 3, 4, 6}, {4, 5, 6}, {4, 6}}
	t.Run("All", testAll)
	wants = [][]int{{0, 1, 2, 3, 4, 6}, {4, 5, 6}, {0, 1, 2, 3, 4, 5, 6}}
	t.Run("Any", testAny)

	c3 := NewComponent(w)
	for i, e := range entities[5:8] {
		SetComp(w, e, c3, i+5)
	}

	// id: [0 1 2 3 4 5 6 7 8 9]
	// c1: [0 1 2 3 4   6      ]
	// c2: [      _ 4 5 6      ]
	// c3: [          5 6 7    ]

	wants = [][]int{{0, 1, 2, 3, 4, 6}, {4, 5, 6}, {4, 6}}
	t.Run("All", testAll)
	wants = [][]int{{0, 1, 2, 3, 4, 6}, {4, 5, 6}, {0, 1, 2, 3, 4, 5, 6}}
	t.Run("Any", testAny)
}

func TestFilter_Cache(t *testing.T) {
	w := NewWorld()
	var entities [10]Entity
	for i := range entities {
		entities[i] = NewEntity(w)
	}

	c1 := NewComponent(w)
	for i, e := range entities[:5] {
		SetComp(w, e, c1, i)
	}
	c2 := NewComponent(w)
	for i, e := range entities[3:7] {
		SetComp(w, e, c2, i+3)
	}

	// id: [0 1 2 3 4 5 6 7 8 9]
	// c1: [0 1 2 3 4          ]
	// c2: [      3 4 5 6      ]

	// Create the cached query
	queryBoth := QueryAll(c1, c2).Cache(w)
	var result []int
	var want []int

	judge := func() {
		result = result[:0]
		queryBoth.Run(func(entities Table[Entity], data []any) {
			result = append(result, []int(*data[0].(*Table[int]))...)
		})
		sort.Ints(result)
		if !reflect.DeepEqual(result, want) {
			t.Errorf("get: %v, want: %v", result, want)
		}
	}

	// Test the basic query
	want = []int{3, 4}
	judge()

	// Test if the cached query gets up to date when entities are moved
	SetComp(w, entities[6], c1, 6)
	DelComp(w, entities[3], c2)

	// id: [0 1 2 3 4 5 6 7 8 9]
	// c1: [0 1 2 3 4   6      ]
	// c2: [      _ 4 5 6      ]

	want = []int{4, 6}
	judge()

	// Test if the cached query gets up to date when a new archetype is created.
	c3 := NewComponent(w)
	for i, e := range entities[5:8] {
		SetComp(w, e, c3, i+5)
	}

	// id: [0 1 2 3 4 5 6 7 8 9]
	// c1: [0 1 2 3 4   6      ]
	// c2: [      _ 4 5 6      ]
	// c3: [          5 6 7    ]

	want = []int{4, 6}
	judge()
}

func BenchmarkFilter_All(b *testing.B) {
	const EntityCount = 1_000_000
	const ComponentCount = 16
	const QueryCount = 3

	w := NewWorld()

	var components [ComponentCount]Component
	for i := range components {
		components[i] = NewComponent(w)
	}

	// count of tables before creating entities
	tableCount := len(w.archetypes)

	for i := 0; i < EntityCount; i++ {
		e := NewEntity(w)
		coins := rand.Int() // we know len(components) < bitsOf(int)
		for i, c := range components {
			if coins&(1<<i) != 0 {
				AddComp(w, e, c)
			}
		}
	}
	b.Logf("entities created: %d (w/%d randomized components)", EntityCount, ComponentCount)
	b.Logf("tables created  : %d", len(w.archetypes)-tableCount)
	b.Logf("setup time      : %v", b.Elapsed())
	b.Logf("queriying for %d components", QueryCount)

	rand.Shuffle(len(components), func(i, j int) {
		components[i], components[j] = components[j], components[i]
	})

	b.Run("uncached", func(b *testing.B) {
		var tableMatched int64
		for i := 0; i < b.N; i++ {
			QueryAll(components[:QueryCount]...).Run(w, func(entities Table[Entity], data []any) {
				tableMatched++
			})
		}
		b.ReportMetric(float64(b.Elapsed().Nanoseconds()/(tableMatched)), "ns/table")
	})
	b.Run("cached", func(b *testing.B) {
		cachedQuery := QueryAll(components[:QueryCount]...).Cache(w)
		b.ResetTimer()

		var tableMatched int64
		for i := 0; i < b.N; i++ {
			cachedQuery.Run(func(entities Table[Entity], data []any) {
				tableMatched++
			})
		}
		b.ReportMetric(float64(b.Elapsed().Nanoseconds()/(tableMatched)), "ns/table")
	})
}
