package ecs

import "testing"

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
