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

	position := ecs.NewComponent(world)
	walking := ecs.NewComponent(world)

	// Create an entity with name Bob
	bob := ecs.NewEntity(world)

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
	alice := ecs.NewEntity(world)
	ecs.Set(world, alice, position, Position{10, 20})
	ecs.Set(world, alice, walking, Walking{})

	// Print all the components the entity has. This will output:
	//    Position, Walking, (Identifier,Name)
	fmt.Printf("[%s]\n", ecs.Type(world, alice))

	// Remove tag
	ecs.Remove[Walking](world, alice, walking)

	// Iterate all entities with Position
	//it := ecs.TermIter[Position](ecs, nil)
	//for it.Next() {
	//	p := ecs.Field[Position](it, 1)
	//	for i := 0; i < len(it.Entities); i++ {
	//		fmt.Printf("%s: {%f, %f}\n", it.Entities[i].Name(ecs), p[i].x, p[i].y)
	//	}
	//}

	// Output:
	// {10.000000, 20.000000}
	// [Position, Walking, (Identifier,Name)]
	// Alice: {10.000000, 20.000000}
	// Bob: {20.000000, 30.000000}
}
