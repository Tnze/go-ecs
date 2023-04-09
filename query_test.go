package ecs

import (
	"fmt"
	"math/rand"
	"testing"
)

func ExampleFilter_All() {
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

	// Current layout:
	//
	// entity:[0 1 2 3 4 5 6 7 8 9]
	// c1:    [0 1 2 3 4          ]
	// c2:    [      3 4 5 6 7    ]
	// c1&c2: [      3 4          ]

	// Query all entities which have both c1 and c2.
	Filter{c1, c2}.All(w, func(entities Table[Entity], data []any) {
		// The type of the data's element is `Table[T]`,
		// which can be converted to `[]T` only after type assertion.
		fmt.Println([]int(*data[0].(*Table[int])))
	})

	// The cached query version.
	cache := Filter{c1, c2}.CacheAll(w)
	cache.Run(func(entities Table[Entity], data []any) {
		fmt.Println([]int(*data[0].(*Table[int])))
	})

	// Output:
	// [3 4]
	// [3 4]
}

func BenchmarkFilter_All(b *testing.B) {
	const EntityCount = 1000_000
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
			Filter(components[:QueryCount]).All(w, func(entities Table[Entity], data []any) {
				tableMatched++
			})
		}
		b.ReportMetric(float64(b.Elapsed().Nanoseconds()/(tableMatched)), "ns/table")
	})
	b.Run("cached", func(b *testing.B) {
		cachedQuery := Filter(components[:QueryCount]).CacheAll(w)
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
