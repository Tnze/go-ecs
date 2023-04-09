package ecs

import "testing"

func TestAddComp_addComponent(t *testing.T) {
	w := NewWorld()
	position := NewComponent(w)
	e := NewEntity(w)

	AddComp(w, e, position)
	if !HasComp(w, e, position) {
		t.Errorf("Unexpected")
	}
}

func TestAddComp_addComponentAgain(t *testing.T) {
	w := NewWorld()
	position := NewComponent(w)
	e := NewEntity(w)

	AddComp(w, e, position)
	if !HasComp(w, e, position) {
		t.Errorf("Unexpected")
	}

	AddComp(w, e, position)
	if !HasComp(w, e, position) {
		t.Errorf("Unexpected")
	}
}

func TestAddComp_add2Component(t *testing.T) {
	w := NewWorld()
	position := NewComponent(w)
	velocity := NewComponent(w)
	e := NewEntity(w)

	AddComp(w, e, position)
	if !HasComp(w, e, position) || HasComp(w, e, velocity) {
		t.Errorf("Unexpected")
	}

	AddComp(w, e, velocity)
	if !HasComp(w, e, position) || !HasComp(w, e, velocity) {
		t.Errorf("Unexpected")
	}
}

func TestAddComp_add2ComponentAgain(t *testing.T) {
	w := NewWorld()
	position := NewComponent(w)
	velocity := NewComponent(w)
	e := NewEntity(w)

	AddComp(w, e, position)
	AddComp(w, e, velocity)
	if !HasComp(w, e, position) || !HasComp(w, e, velocity) {
		t.Errorf("Unexpected")
	}

	AddComp(w, e, position)
	AddComp(w, e, velocity)
	if !HasComp(w, e, position) || !HasComp(w, e, velocity) {
		t.Errorf("Unexpected")
	}
}

func TestAddComp_add2ComponentOverlap(t *testing.T) {
	w := NewWorld()
	position := NewComponent(w)
	velocity := NewComponent(w)
	mass := NewComponent(w)
	e := NewEntity(w)

	AddComp(w, e, position)
	AddComp(w, e, velocity)
	if !HasComp(w, e, position) || !HasComp(w, e, velocity) || HasComp(w, e, mass) {
		t.Errorf("Unexpected")
	}

	AddComp(w, e, velocity)
	AddComp(w, e, mass)
	if !HasComp(w, e, position) || !HasComp(w, e, velocity) || !HasComp(w, e, mass) {
		t.Errorf("Unexpected")
	}
}
