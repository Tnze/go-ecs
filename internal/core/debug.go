package core

import (
	"sort"
	"strings"
)

func Type(w *World, e Entity, nameComp Component) string {
	var sb strings.Builder
	rec := w.Entities[e]
	compNames := make([]string, len(rec.AT.Types))
	for i, v := range rec.AT.Types {
		switch name := GetComp[string](w, v.Entity, nameComp); name {
		case nil:
			// type of v.TableType has to be `*Table[T]` which .Elem is `Table[T]` which .Elem is `T`
			compNames[i] = v.TableType.Elem().Elem().Name()
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
