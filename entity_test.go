package ecs_test

import (
	"fmt"

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
	ecs.Remove[Walking](world, alice, walking)

	// Output:
	// {10.000000, 20.000000}
	// [Position, Walking, ecs.Name]
	// Bob: {20.000000, 30.000000}
	// Alice: {10.000000, 20.000000}
}
