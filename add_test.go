package ecs

import (
	"testing"

	"github.com/Tnze/go-ecs/internal/core"
)

func TestAddComp_addComponent(t *testing.T) {
	w := core.NewWorld()
	position := core.NewComponent(w)
	e := core.NewEntity(w)

	core.AddComp(w, e, position)
	if !core.HasComp(w, e, position) {
		t.FailNow()
	}
}

func TestAddComp_addComponentAgain(t *testing.T) {
	w := core.NewWorld()
	position := core.NewComponent(w)
	e := core.NewEntity(w)

	core.AddComp(w, e, position)
	if !core.HasComp(w, e, position) {
		t.FailNow()
	}

	core.AddComp(w, e, position)
	if !core.HasComp(w, e, position) {
		t.FailNow()
	}
}

func TestAddComp_add2Component(t *testing.T) {
	w := core.NewWorld()
	position := core.NewComponent(w)
	velocity := core.NewComponent(w)
	e := core.NewEntity(w)

	core.AddComp(w, e, position)
	if !core.HasComp(w, e, position) || core.HasComp(w, e, velocity) {
		t.FailNow()
	}

	core.AddComp(w, e, velocity)
	if !core.HasComp(w, e, position) || !core.HasComp(w, e, velocity) {
		t.FailNow()
	}
}

func TestAddComp_add2ComponentAgain(t *testing.T) {
	w := core.NewWorld()
	position := core.NewComponent(w)
	velocity := core.NewComponent(w)
	e := core.NewEntity(w)

	core.AddComp(w, e, position)
	core.AddComp(w, e, velocity)
	if !core.HasComp(w, e, position) || !core.HasComp(w, e, velocity) {
		t.FailNow()
	}

	core.AddComp(w, e, position)
	core.AddComp(w, e, velocity)
	if !core.HasComp(w, e, position) || !core.HasComp(w, e, velocity) {
		t.FailNow()
	}
}

func TestAddComp_add2ComponentOverlap(t *testing.T) {
	w := core.NewWorld()
	position := core.NewComponent(w)
	velocity := core.NewComponent(w)
	mass := core.NewComponent(w)
	e := core.NewEntity(w)

	core.AddComp(w, e, position)
	core.AddComp(w, e, velocity)
	if !core.HasComp(w, e, position) || !core.HasComp(w, e, velocity) || core.HasComp(w, e, mass) {
		t.FailNow()
	}

	core.AddComp(w, e, velocity)
	core.AddComp(w, e, mass)
	if !core.HasComp(w, e, position) || !core.HasComp(w, e, velocity) || !core.HasComp(w, e, mass) {
		t.FailNow()
	}
}
