package ecs

import (
	"fmt"
	"testing"
)

func TestEntity_basic(t *testing.T) {
	w := NewWorld()
	name := w.NewComponent()
	c1 := w.NewComponent()
	c2 := w.NewComponent()
	e1 := w.NewEntity()
	e2 := w.NewEntity()
	e3 := w.NewEntity()
	w.SetComp(e1, name, "E1")
	w.SetComp(e2, name, "E2")
	w.SetComp(e3, name, "E3")
	w.SetComp(e1, c1, "E1-C1")
	w.SetComp(e2, c1, "E2-C1")
	w.SetComp(e2, c2, "E2-C2")
	w.SetComp(e3, c2, "E2-C2")

	w.Query(QueryAll(c1), func(entities []Entity, data []any) {
		s := *data[0].(*[]string)
		for i, e := range entities {
			entityName := w.GetComp[string](e, name)
			fmt.Printf("%s: %s\n", *entityName, s[i])
		}
	})
	w.Query(QueryAll(c2), func(entities []Entity, data []any) {
		s := *data[0].(*[]string)
		for i, e := range entities {
			entityName := w.GetComp[string](e, name)
			fmt.Printf("%s: %s\n", *entityName, s[i])
		}
	})
}

func TestNewEntity(t *testing.T) {
	w := NewWorld()

	// Test create entity
	var entities [10]Entity
	for i := range entities {
		entities[i] = w.NewEntity()
	}

	// Test recycle ids
	for i := range entities {
		w.DelEntity(entities[i])
	}
	for i := range entities {
		entities[i] = w.NewEntity()
	}

	// We create 10 entities, and delete 10 entities, and then create 10 entities again.
	// The latter 10 entities should reuse the ids of the former 10 entities.
	// So NextID should not be 20 but 10.
	if w.IDManager.NextID >= 20 {
		t.Errorf("idManager doesn't recycle ids")
	}
}

func TestDelEntity(t *testing.T) {
	w := NewWorld()

	var entities [100]Entity
	for i := range entities {
		entities[i] = w.NewEntity()
	}

	for i := range entities {
		w.DelEntity(entities[i])
	}
}

func TestDelComp(t *testing.T) {
	w := NewWorld()
	e := w.NewEntity()

	var components [100]Component
	for i := range components {
		components[i] = w.NewComponent()
	}

	for j := range components {
		w.SetComp(e, components[j], j)
	}

	for j := range components {
		w.DelComp(e, components[j])
	}
}

func BenchmarkNewEntity(b *testing.B) {
	w := NewWorld()
	for b.Loop() {
		w.NewEntity()
	}
}

func BenchmarkAddComp_millionEntities(b *testing.B) {
	prepare := func(n int) (w *World, entities []Entity, components []Component) {
		w = NewWorld()
		entities = make([]Entity, 1_000_000)
		for i := range entities {
			entities[i] = w.NewEntity()
		}
		components = make([]Component, n)
		for i := range components {
			components[i] = w.NewComponent()
		}
		return
	}

	b.Run("ByHash", func(b *testing.B) {
		w, entities, components := prepare(b.N)

		b.ResetTimer()

		for i, e := range entities {
			for c := i; c < b.N+i; c++ {
				w.AddComp(e, components[c%b.N])
			}
		}
	})

	b.Run("ByShortcuts", func(b *testing.B) {
		w, entities, components := prepare(b.N)
		// create shortcuts
		tmpEntity := w.NewEntity()
		for c := 0; c < b.N; c++ {
			w.AddComp(tmpEntity, components[c])
		}

		b.ResetTimer()

		for _, e := range entities {
			for c := 0; c < b.N; c++ {
				w.AddComp(e, components[c])
			}
		}
	})
}
