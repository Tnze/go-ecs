package ecs

import (
	"testing"
)

func TestAddComp_addComponent(t *testing.T) {
	w := NewWorld()
	position := w.NewComponent()
	e := w.NewEntity()

	w.AddComp(e, position)
	if !w.HasComp(e, position) {
		t.FailNow()
	}
}

func TestAddComp_addComponentAgain(t *testing.T) {
	w := NewWorld()
	position := w.NewComponent()
	e := w.NewEntity()

	w.AddComp(e, position)
	if !w.HasComp(e, position) {
		t.FailNow()
	}

	w.AddComp(e, position)
	if !w.HasComp(e, position) {
		t.FailNow()
	}
}

func TestAddComp_add2Component(t *testing.T) {
	w := NewWorld()
	position := w.NewComponent()
	velocity := w.NewComponent()
	e := w.NewEntity()

	w.AddComp(e, position)
	if !w.HasComp(e, position) || w.HasComp(e, velocity) {
		t.FailNow()
	}

	w.AddComp(e, velocity)
	if !w.HasComp(e, position) || !w.HasComp(e, velocity) {
		t.FailNow()
	}
}

func TestAddComp_add2ComponentAgain(t *testing.T) {
	w := NewWorld()
	position := w.NewComponent()
	velocity := w.NewComponent()
	e := w.NewEntity()

	w.AddComp(e, position)
	w.AddComp(e, velocity)
	if !w.HasComp(e, position) || !w.HasComp(e, velocity) {
		t.FailNow()
	}

	w.AddComp(e, position)
	w.AddComp(e, velocity)
	if !w.HasComp(e, position) || !w.HasComp(e, velocity) {
		t.FailNow()
	}
}

func TestAddComp_add2ComponentOverlap(t *testing.T) {
	w := NewWorld()
	position := w.NewComponent()
	velocity := w.NewComponent()
	mass := w.NewComponent()
	e := w.NewEntity()

	w.AddComp(e, position)
	w.AddComp(e, velocity)
	if !w.HasComp(e, position) || !w.HasComp(e, velocity) || w.HasComp(e, mass) {
		t.FailNow()
	}

	w.AddComp(e, velocity)
	w.AddComp(e, mass)
	if !w.HasComp(e, position) || !w.HasComp(e, velocity) || !w.HasComp(e, mass) {
		t.FailNow()
	}
}
