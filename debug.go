package ecs

import (
	"sort"
	"strconv"
	"strings"
)

func Type(w *World, e Entity, nameComp Component) string {
	var sb strings.Builder
	rec := w.entities[e]
	compNames := make([]string, len(rec.at.types))
	for i, v := range rec.at.types {
		switch name := Get[string](w, v.Entity, nameComp); name {
		case nil:
			compNames[i] = "<unnamed(" + strconv.FormatUint(uint64(v.Entity), 10) + ")>"
		default:
			compNames[i] = *name
		}
	}
	sort.Strings(compNames)
	for i, v := range compNames {
		if i != 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(v)
	}
	return sb.String()
}
