package ecs

import "fmt"

func ExampleFilter_All() {
	w := NewWorld()

	var entities [100]Entity
	for i := range entities {
		entities[i] = NewEntity(w)
	}

	c1 := NewComponent(w)
	c2 := NewComponent(w)
	for _, e := range entities[:50] {
		AddComp(w, e, c1)
	}
	for i, e := range entities[30:70] {
		SetComp(w, e, c2, i)
	}

	Filter{c1, c2}.All(w, func(entities []Entity, data []any) {
		fmt.Println(entities)
		// len(data) == 1 because c1 doesn't store data
		fmt.Println([]int(*data[0].(*Table[int])))
	})

	// Output:
	// [30 31 32 33 34 35 36 37 38 39 40 41 42 43 44 45 46 47 48 49]
	// [0 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19]
}
