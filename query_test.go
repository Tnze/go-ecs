package ecs

import "fmt"

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
	Filter{c1, c2}.All(w, func(entities []Entity, data []any) {
		// The type of the data's element is `Table[T]`,
		// which can be converted to `[]T` only after type assertion.
		fmt.Println([]int(*data[0].(*Table[int])))
	})

	// Output:
	// [3 4]
}
