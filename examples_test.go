package ecs_test

import (
	"fmt"
	"sort"

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
	pos := ecs.GetComp[Position](w, bob, position)
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
			entityName := ecs.GetComp[string](w, e, name)
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

func ExampleQueryAll() {
	w := ecs.NewWorld()

	// Create 10 entities.
	var entities [10]ecs.Entity
	for i := range entities {
		entities[i] = ecs.NewEntity(w)
	}

	// Create 2 components.
	c1 := ecs.NewComponent(w)
	c2 := ecs.NewComponent(w)

	// Add components to entities.
	for i, e := range entities[:5] {
		ecs.SetComp(w, e, c1, i)
	}
	for i, e := range entities[3:7] {
		ecs.SetComp(w, e, c2, i+3)
	}

	// Current layout:
	//
	// entity:[0 1 2 3 4 5 6 7 8 9]
	// c1:    [0 1 2 3 4          ]
	// c2:    [      3 4 5 6      ]
	// c1&c2: [      3 4          ]

	// CachedQuery all entities which have both c1 and c2.
	ecs.QueryAll(c1, c2).Run(w, func(entities ecs.Table[ecs.Entity], data []any) {
		// The type of the data's element is `Table[T]`,
		// which can be converted to `[]T` only after type assertion.
		fmt.Println([]int(*data[0].(*ecs.Table[int])))
	})

	// Output:
	// [3 4]
}

func ExampleQueryAny() {
	w := ecs.NewWorld()

	// Create 10 entities.
	var entities [10]ecs.Entity
	for i := range entities {
		entities[i] = ecs.NewEntity(w)
	}

	// Create 2 components.
	c1 := ecs.NewComponent(w)
	c2 := ecs.NewComponent(w)

	// Add components to entities.
	for i, e := range entities[:5] {
		ecs.SetComp(w, e, c1, i)
	}
	for i, e := range entities[3:7] {
		ecs.SetComp(w, e, c2, i+3)
	}

	// Current layout:
	//
	// entity:[0 1 2 3 4 5 6 7 8 9]
	// c1:    [0 1 2 3 4          ]
	// c2:    [      3 4 5 6      ]
	// c1&c2: [      3 4          ]

	// CachedQuery all entities which have both c1 and c2.
	var result []int
	ecs.QueryAny(c1, c2).Run(w, func(entities ecs.Table[ecs.Entity], data []any) {
		// The type of the data's element is `Table[T]`,
		// which can be converted to `[]T` only after type assertion.
		result = append(result, *data[0].(*ecs.Table[int])...)
	})
	sort.Ints(result)
	fmt.Println(result)

	// Output:
	// [0 1 2 3 4 5 6]
}
