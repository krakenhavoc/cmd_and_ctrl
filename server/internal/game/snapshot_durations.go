package game

import (
	"encoding"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"
)

// snapshot_durations.go is ADR 0109 Shared machinery 2: the restore
// check on EVERY stored duration.
//
// A Duration is data, and it is stored in more places than the
// scopedEffects registry: a delayed trigger's lifetime, a granted cast
// permission's, a player static's, and an untap hold's `while` on any
// card in any zone. Its condition is a bare int and its newer fields
// (CounterKind, Also) are keys a binary before them drops without a
// word. Before this file only scopedEffects and delayedTriggers were
// checked, so an older binary read a counter-held untap hold as "while
// the source is on the battlefield".
//
// The check is derived from the snapshot TYPES, not from a list of
// places: durationRoutes walks GameSnapshot by reflection, the way the
// shape guard does, and finds every route to a Duration. A new field
// that stores one is covered by the check the day it is added. Two
// halves, both failing closed:
//
//   - unknownDurationFields re-reads the raw file along those routes
//     and lists every key on a duration this binary does not declare
//     (encoding/json would drop it). Called from UnmarshalJSON, like
//     unknownScopedEffectFields.
//   - unknownDurations walks the decoded snapshot along the same routes
//     and asks Duration.Problem of every value. Called from
//     checkEffectKeys, so a snapshot built in memory is checked too.
//
// Both are refused as ErrUnknownEffectKey: the rollback case, keep the
// file (ADR 0041 P4).

var (
	durationType   = reflect.TypeOf(Duration{})
	jsonMarshalerT = reflect.TypeOf((*json.Marshaler)(nil)).Elem()
	textMarshalerT = reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()

	durationReachOnce sync.Once
	durationReach     map[reflect.Type]bool
)

// reachesDuration reports whether a value of type t can hold a Duration
// in the JSON encoding/json writes for it. Computed once, as a fixed
// point over every type reachable from GameSnapshot, so a recursive
// type is answered correctly whichever way round the cycle is entered.
// A type outside that graph answers false.
func reachesDuration(t reflect.Type) bool {
	durationReachOnce.Do(buildDurationReach)
	return durationReach[derefType(t)]
}

func derefType(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t
}

// jsonChildren is the types one level inside t, as encoding/json sees
// it. A type with its own marshaler has none: its fields are not its
// JSON.
func jsonChildren(t reflect.Type) []reflect.Type {
	if t == durationType || customJSON(t) {
		return nil
	}
	switch t.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map:
		return []reflect.Type{derefType(t.Elem())}
	case reflect.Struct:
		var out []reflect.Type
		for _, f := range jsonFields(t) {
			out = append(out, derefType(f.typ))
		}
		return out
	}
	return nil
}

func buildDurationReach() {
	children := map[reflect.Type][]reflect.Type{}
	queue := []reflect.Type{reflect.TypeOf(GameSnapshot{})}
	for len(queue) > 0 {
		t := queue[0]
		queue = queue[1:]
		if _, seen := children[t]; seen {
			continue
		}
		children[t] = jsonChildren(t)
		queue = append(queue, children[t]...)
	}
	reach := map[reflect.Type]bool{durationType: true}
	for changed := true; changed; {
		changed = false
		for t, kids := range children {
			if reach[t] {
				continue
			}
			for _, k := range kids {
				if reach[k] {
					reach[t], changed = true, true
					break
				}
			}
		}
	}
	durationReach = reach
}

// customJSON reports whether t writes its own JSON, which reflection
// over its fields does not describe.
func customJSON(t reflect.Type) bool {
	return t.Implements(jsonMarshalerT) || reflect.PointerTo(t).Implements(jsonMarshalerT) ||
		t.Implements(textMarshalerT) || reflect.PointerTo(t).Implements(textMarshalerT)
}

// jsonField is one key encoding/json writes for a struct, with the
// index path to the Go field (more than one step for a field promoted
// from an untagged embedded struct).
type jsonField struct {
	name  string
	index []int
	typ   reflect.Type
}

// jsonFields lists the keys encoding/json writes for struct type t,
// flattening untagged embedded structs as it does.
func jsonFields(t reflect.Type) []jsonField {
	var out []jsonField
	var walk func(t reflect.Type, prefix []int)
	walk = func(t reflect.Type, prefix []int) {
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			tag := f.Tag.Get("json")
			if tag == "-" {
				continue
			}
			name, _, _ := strings.Cut(tag, ",")
			index := append(append([]int(nil), prefix...), i)
			if f.Anonymous && name == "" {
				ft := f.Type
				for ft.Kind() == reflect.Pointer {
					ft = ft.Elem()
				}
				if ft.Kind() == reflect.Struct {
					walk(ft, index)
					continue
				}
			}
			if !f.IsExported() {
				continue
			}
			if name == "" {
				name = f.Name
			}
			out = append(out, jsonField{name: name, index: index, typ: f.Type})
		}
	}
	walk(t, nil)
	return out
}

// durationRoutes lists every JSON path in a GameSnapshot at which a
// Duration is stored, for tests and for the operator reading a refusal.
// Array elements are "[]" and map values "{}"; a recursive type is
// followed one level, as the shape guard does.
func durationRoutes() []string {
	set := map[string]bool{}
	var walk func(t reflect.Type, path string, stack []reflect.Type)
	walk = func(t reflect.Type, path string, stack []reflect.Type) {
		for t.Kind() == reflect.Pointer {
			t = t.Elem()
		}
		if t == durationType {
			set[path] = true
			return
		}
		if !reachesDuration(t) {
			return
		}
		switch t.Kind() {
		case reflect.Slice, reflect.Array:
			walk(t.Elem(), path+"[]", stack)
		case reflect.Map:
			walk(t.Elem(), path+"{}", stack)
		case reflect.Struct:
			for _, s := range stack {
				if s == t {
					return
				}
			}
			for _, f := range jsonFields(t) {
				child := f.name
				if path != "" {
					child = path + "." + f.name
				}
				walk(f.typ, child, append(stack, t))
			}
		}
	}
	walk(reflect.TypeOf(GameSnapshot{}), "", nil)
	out := make([]string, 0, len(set))
	for p := range set {
		out = append(out, p)
	}
	sort.Strings(out)
	return out
}

// unknownDurationFields re-reads the raw snapshot along every route to
// a Duration and lists each key this binary's Duration does not
// declare. Asked of every schema: no Duration key has ever been removed
// or renamed, so an unknown one can only be a newer binary's.
func unknownDurationFields(data []byte) ([]string, error) {
	var out []string
	err := walkRawDurations(data, reflect.TypeOf(GameSnapshot{}), "", func(where string, obj map[string]json.RawMessage) {
		for k := range obj {
			if !durationJSONKeys[k] {
				out = append(out, fmt.Sprintf("unknown field %q on a duration at %s", k, where))
			}
		}
	})
	sort.Strings(out)
	return out, err
}

// walkRawDurations follows type t through raw JSON, calling visit with
// the decoded keys of every Duration object it meets. Only the branches
// that can hold a Duration are decoded.
func walkRawDurations(raw json.RawMessage, t reflect.Type, where string, visit func(string, map[string]json.RawMessage)) error {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if len(raw) == 0 || string(raw) == "null" || !reachesDuration(t) {
		return nil
	}
	if t == durationType {
		var obj map[string]json.RawMessage
		if err := json.Unmarshal(raw, &obj); err != nil {
			return err
		}
		visit(where, obj)
		return nil
	}
	switch t.Kind() {
	case reflect.Slice, reflect.Array:
		var items []json.RawMessage
		if err := json.Unmarshal(raw, &items); err != nil {
			return err
		}
		for _, it := range items {
			if err := walkRawDurations(it, t.Elem(), where+"[]", visit); err != nil {
				return err
			}
		}
	case reflect.Map:
		var items map[string]json.RawMessage
		if err := json.Unmarshal(raw, &items); err != nil {
			return err
		}
		for _, it := range items {
			if err := walkRawDurations(it, t.Elem(), where+"{}", visit); err != nil {
				return err
			}
		}
	case reflect.Struct:
		var obj map[string]json.RawMessage
		if err := json.Unmarshal(raw, &obj); err != nil {
			return err
		}
		for _, f := range jsonFields(t) {
			child := f.name
			if where != "" {
				child = where + "." + f.name
			}
			if err := walkRawDurations(obj[f.name], f.typ, child, visit); err != nil {
				return err
			}
		}
	}
	return nil
}

// unknownDurations walks the decoded snapshot along every route to a
// Duration and lists each one this binary cannot interpret
// (Duration.Problem), naming where it was.
func (s *GameSnapshot) unknownDurations() []string {
	var out []string
	walkDurationValues(reflect.ValueOf(s), "", func(where string, d Duration) {
		if problem := d.Problem(); problem != "" {
			out = append(out, fmt.Sprintf("%s at %s", problem, where))
		}
	})
	sort.Strings(out)
	return out
}

func walkDurationValues(v reflect.Value, where string, visit func(string, Duration)) {
	for v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}
	t := v.Type()
	if t == durationType {
		visit(where, v.Interface().(Duration))
		return
	}
	if !reachesDuration(t) {
		return
	}
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			walkDurationValues(v.Index(i), where+"[]", visit)
		}
	case reflect.Map:
		iter := v.MapRange()
		for iter.Next() {
			walkDurationValues(iter.Value(), where+"{}", visit)
		}
	case reflect.Struct:
		for _, f := range jsonFields(t) {
			child := f.name
			if where != "" {
				child = where + "." + f.name
			}
			fv, err := v.FieldByIndexErr(f.index)
			if err != nil {
				continue
			}
			walkDurationValues(fv, child, visit)
		}
	}
}
