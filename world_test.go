package ecs

import (
	"fmt"
	"testing"
)

func TestEntity_basic(t *testing.T) {
	w := NewWorld()
	name := NewComponent(w)
	c1 := NewComponent(w)
	c2 := NewComponent(w)
	e1 := NewEntity(w)
	e2 := NewEntity(w)
	e3 := NewEntity(w)
	SetComp(w, e1, name, "E1")
	SetComp(w, e2, name, "E2")
	SetComp(w, e3, name, "E3")
	SetComp(w, e1, c1, "E1-C1")
	SetComp(w, e2, c1, "E2-C1")
	SetComp(w, e2, c2, "E2-C2")
	SetComp(w, e3, c2, "E2-C2")

	QueryAll(c1).Run(w, func(entities Table[Entity], data []any) {
		s := *data[0].(*Table[string])
		for i, e := range entities {
			entityName := Get[string](w, e, name)
			fmt.Printf("%s: %s\n", *entityName, s[i])
		}
	})
	QueryAll(c2).Run(w, func(entities Table[Entity], data []any) {
		s := *data[0].(*Table[string])
		for i, e := range entities {
			entityName := Get[string](w, e, name)
			fmt.Printf("%s: %s\n", *entityName, s[i])
		}
	})
}

func TestNewEntity(t *testing.T) {
	w := NewWorld()

	// Test create entity
	var entities [10]Entity
	for i := range entities {
		entities[i] = NewEntity(w)
	}

	// Test recycle ids
	for i := range entities {
		DelEntity(w, entities[i])
	}
	for i := range entities {
		entities[i] = NewEntity(w)
	}

	// We create 10 entities, and delete 10 entities, and then create 10 entities again.
	// The latter 10 entities should reuse the ids of the former 10 entities.
	// So nextID should not be 20 but 10.
	if w.idManager.nextID >= 20 {
		t.Errorf("idManager doesn't recycle ids")
	}
}

func TestDelEntity(t *testing.T) {
	w := NewWorld()

	var entities [100]Entity
	for i := range entities {
		entities[i] = NewEntity(w)
	}

	for i := range entities {
		DelEntity(w, entities[i])
	}
}

func TestDelComp(t *testing.T) {
	w := NewWorld()
	e := NewEntity(w)

	var components [100]Component
	for i := range components {
		components[i] = NewComponent(w)
	}

	for j := range components {
		SetComp(w, e, components[j], j)
	}

	for j := range components {
		DelComp(w, e, components[j])
	}
}

func BenchmarkNewEntity(b *testing.B) {
	w := NewWorld()
	for i := 0; i < b.N; i++ {
		NewEntity(w)
	}
}

func BenchmarkAddComp_millionEntities(b *testing.B) {
	prepare := func(n int) (w *World, entities []Entity, components []Component) {
		w = NewWorld()
		entities = make([]Entity, 1_000_000)
		for i := range entities {
			entities[i] = NewEntity(w)
		}
		components = make([]Component, n)
		for i := range components {
			components[i] = NewComponent(w)
		}
		return
	}

	b.Run("ByHash", func(b *testing.B) {
		w, entities, components := prepare(b.N)

		b.ResetTimer()

		for i, e := range entities {
			for c := i; c < b.N+i; c++ {
				AddComp(w, e, components[c%b.N])
			}
		}
	})

	b.Run("ByShortcuts", func(b *testing.B) {
		w, entities, components := prepare(b.N)
		// create shortcuts
		tmpEntity := NewEntity(w)
		for c := 0; c < b.N; c++ {
			AddComp(w, tmpEntity, components[c])
		}

		b.ResetTimer()

		for _, e := range entities {
			for c := 0; c < b.N; c++ {
				AddComp(w, e, components[c])
			}
		}
	})
}
