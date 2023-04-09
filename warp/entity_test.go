package warp_test

import (
	"fmt"

	ecs "github.com/Tnze/go-ecs/warp"
)

func ExampleEntity_basic() {
	type (
		Position struct{ x, y float64 }
		Walking  struct{}
	)

	w := ecs.NewWorld()

	// Create an entity with name Bob
	bob := w.NewNamedEntity("Bob")

	// The set operation finds or creates a component, and sets it.
	ecs.SetComp(bob, Position{10, 20})
	// The add operation adds a component without setting a value. This is
	// useful for tags, or when adding a component with its default value.
	ecs.SetComp(bob, Walking{})

	// Get the value for the Position component
	pos := ecs.GetComp[Position](bob)
	fmt.Printf("{%f, %f}\n", pos.x, pos.y)

	// Overwrite the value of the Position component
	ecs.SetComp(bob, Position{20, 30})

	// Create another named entity
	alice := w.NewNamedEntity("Alice")
	ecs.SetComp(alice, Position{10, 20})
	ecs.SetComp(alice, Walking{})

	// Print all the components the entity has. This will output:
	//    Position, Walking, (Identifier,Name)
	fmt.Printf("[%s]\n", alice.Type())
	// Iterate all entities with Position

	w.Filter(func(entities []ecs.Entity, positions []Position) {
		for i, e := range entities {
			fmt.Printf("%s: {%f, %f}\n", *e.Name(), positions[i].x, positions[i].y)
		}
	})
	// DelComp tag
	ecs.DelComp[Walking](alice)

	// Output:
	// {10.000000, 20.000000}
	// [Name, Position, Walking]
	// Bob: {20.000000, 30.000000}
	// Alice: {10.000000, 20.000000}
}
