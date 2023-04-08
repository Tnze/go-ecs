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

	world := ecs.NewWorld()

	position := ecs.NewNamedComponent(world, "Position")
	walking := ecs.NewNamedComponent(world, "Walking")

	// Create an entity with name Bob
	bob := ecs.NewNamedEntity(world, "Bob")

	// The set operation finds or creates a component, and sets it.
	ecs.Set(world, bob, position, Position{10, 20})
	// The add operation adds a component without setting a value. This is
	// useful for tags, or when adding a component with its default value.
	ecs.Set(world, bob, walking, Walking{})

	// Get the value for the Position component
	pos := ecs.Get[Position](world, bob, position)
	fmt.Printf("{%f, %f}\n", pos.x, pos.y)

	// Overwrite the value of the Position component
	ecs.Set(world, bob, position, Position{20, 30})

	// Create another named entity
	alice := ecs.NewNamedEntity(world, "Alice")
	ecs.Set(world, alice, position, Position{10, 20})
	ecs.Set(world, alice, walking, Walking{})

	// Print all the components the entity has. This will output:
	//    Position, Walking, (Identifier,Name)
	fmt.Printf("[%s]\n", ecs.Type(world, alice))
	// Iterate all entities with Position
	ecs.TermIter[Position](world, position, func(entities []ecs.Entity, p []Position) {
		for i, e := range entities {
			entityName := ecs.Get[string](world, e, world.NameComp)
			fmt.Printf("%s: {%f, %f}\n", *entityName, p[i].x, p[i].y)
		}
	})
	// Remove tag
	ecs.Remove(world, alice, walking)

	// Output:
	// {10.000000, 20.000000}
	// [Position, Walking, ecs.Name]
	// Bob: {20.000000, 30.000000}
	// Alice: {10.000000, 20.000000}
}

func TestEntity_basic(t *testing.T) {
	w := ecs.NewWorld()
	c1 := ecs.NewComponent(w)
	c2 := ecs.NewComponent(w)
	e1 := ecs.NewNamedEntity(w, "E1")
	e2 := ecs.NewNamedEntity(w, "E2")
	e3 := ecs.NewNamedEntity(w, "E3")
	ecs.Set(w, e1, c1, "E1-C1")
	ecs.Set(w, e2, c1, "E2-C1")
	ecs.Set(w, e2, c2, "E2-C2")
	ecs.Set(w, e3, c2, "E2-C2")

	ecs.TermIter[string](w, c1, func(entities []ecs.Entity, s []string) {
		for i, e := range entities {
			entityName := ecs.Get[string](w, e, w.NameComp)
			fmt.Printf("%s: %s\n", *entityName, s[i])
		}
	})
	ecs.TermIter[string](w, c2, func(entities []ecs.Entity, s []string) {
		for i, e := range entities {
			entityName := ecs.Get[string](w, e, w.NameComp)
			fmt.Printf("%s: %s\n", *entityName, s[i])
		}
	})
}
