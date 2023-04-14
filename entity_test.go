package ecs_test

import (
	"fmt"
	"testing"

	"github.com/Tnze/go-ecs"
)

func ExampleEntity_basic() {
	type (
		Position struct{ x, y float64 }
		Walking  struct{}
	)

	w := ecs.NewWorld()

	name := ecs.NewComponent(w)
	ecs.SetComp(w, name.Entity, name, "Name")

	position := ecs.NewComponent(w)
	ecs.SetComp(w, position.Entity, name, "Position")

	walking := ecs.NewComponent(w)
	ecs.SetComp(w, walking.Entity, name, "Walking")

	// Create an entity with name Bob
	bob := ecs.NewEntity(w)
	ecs.SetComp(w, bob, name, "Bob")

	// The set operation finds or creates a component, and sets it.
	ecs.SetComp(w, bob, position, Position{10, 20})
	// The add operation adds a component without setting a value. This is
	// useful for tags, or when adding a component with its default value.
	ecs.SetComp(w, bob, walking, Walking{})

	// Get the value for the Position component
	pos := ecs.Get[Position](w, bob, position)
	fmt.Printf("{%f, %f}\n", pos.x, pos.y)

	// Overwrite the value of the Position component
	ecs.SetComp(w, bob, position, Position{20, 30})

	// Create another named entity
	alice := ecs.NewEntity(w)
	ecs.SetComp(w, alice, name, "Alice")
	ecs.SetComp(w, alice, position, Position{10, 20})
	ecs.SetComp(w, alice, walking, Walking{})

	// Print all the components the entity has. This will output:
	//    Position, Walking, (Identifier,Name)
	fmt.Printf("[%s]\n", ecs.Type(w, alice, name))
	// Iterate all entities with Position
	ecs.QueryAll(position).Run(w, func(entities ecs.Table[ecs.Entity], data []any) {
		p := *data[0].(*ecs.Table[Position])
		for i, e := range entities {
			entityName := ecs.Get[string](w, e, name)
			fmt.Printf("%s: {%f, %f}\n", *entityName, p[i].x, p[i].y)
		}
	})
	// DelComp tag
	ecs.DelComp(w, alice, walking)

	// Output:
	// {10.000000, 20.000000}
	// [Name, Position, Walking]
	// Bob: {20.000000, 30.000000}
	// Alice: {10.000000, 20.000000}
}

func TestEntity_basic(t *testing.T) {
	w := ecs.NewWorld()
	name := ecs.NewComponent(w)
	c1 := ecs.NewComponent(w)
	c2 := ecs.NewComponent(w)
	e1 := ecs.NewEntity(w)
	e2 := ecs.NewEntity(w)
	e3 := ecs.NewEntity(w)
	ecs.SetComp(w, e1, name, "E1")
	ecs.SetComp(w, e2, name, "E2")
	ecs.SetComp(w, e3, name, "E3")
	ecs.SetComp(w, e1, c1, "E1-C1")
	ecs.SetComp(w, e2, c1, "E2-C1")
	ecs.SetComp(w, e2, c2, "E2-C2")
	ecs.SetComp(w, e3, c2, "E2-C2")

	ecs.QueryAll(c1).Run(w, func(entities ecs.Table[ecs.Entity], data []any) {
		s := *data[0].(*ecs.Table[string])
		for i, e := range entities {
			entityName := ecs.Get[string](w, e, name)
			fmt.Printf("%s: %s\n", *entityName, s[i])
		}
	})
	ecs.QueryAll(c2).Run(w, func(entities ecs.Table[ecs.Entity], data []any) {
		s := *data[0].(*ecs.Table[string])
		for i, e := range entities {
			entityName := ecs.Get[string](w, e, name)
			fmt.Printf("%s: %s\n", *entityName, s[i])
		}
	})
}

func BenchmarkNewEntity(b *testing.B) {
	w := ecs.NewWorld()
	for i := 0; i < b.N; i++ {
		ecs.NewEntity(w)
	}
}

func BenchmarkAddComp_millionEntities(b *testing.B) {
	prepare := func(n int) (w *ecs.World, entities []ecs.Entity, components []ecs.Component) {
		w = ecs.NewWorld()
		entities = make([]ecs.Entity, 1_000_000)
		for i := range entities {
			entities[i] = ecs.NewEntity(w)
		}
		components = make([]ecs.Component, n)
		for i := range components {
			components[i] = ecs.NewComponent(w)
		}
		return
	}

	b.Run("ByHash", func(b *testing.B) {
		w, entities, components := prepare(b.N)

		b.ResetTimer()

		for i, e := range entities {
			for c := i; c < b.N+i; c++ {
				ecs.AddComp(w, e, components[c%b.N])
			}
		}
	})

	b.Run("ByShortcuts", func(b *testing.B) {
		w, entities, components := prepare(b.N)
		// create shortcuts
		tmpEntity := ecs.NewEntity(w)
		for c := 0; c < b.N; c++ {
			ecs.AddComp(w, tmpEntity, components[c])
		}

		b.ResetTimer()

		for _, e := range entities {
			for c := 0; c < b.N; c++ {
				ecs.AddComp(w, e, components[c])
			}
		}
	})
}
