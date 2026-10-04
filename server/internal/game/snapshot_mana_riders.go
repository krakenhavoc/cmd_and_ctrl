package game

import (
	"fmt"
	"reflect"
	"sort"
	"sync"
)

// snapshot_mana_riders.go — the restore-time refusal of a mana spend
// rider this binary cannot read (ADR 0109 §11, #1552; ADR 0041 P4).
//
// A rider is data on a token (mana_spend_rider.go), and a token lives in
// four kinds of place a restore point carries: a player's pool, a
// pending colour pick, a stack item's payment record and a permanent's
// provenance. Until ADR 0109 PR 10 nothing checked what kind a restored
// rider named, so a token minted by a newer binary — a rider kind added
// after this one, or a counted rider whose count key it does not
// register — would have been spent here as if the rider were not there:
// a Generator Servant's {C}{C} paying for a creature that never gained
// haste. Refusing the file (ErrUnknownEffectKey, the file kept) is the
// rollback answer every other effect key gets.
//
// The walk is by reflection over the snapshot's types, like the
// duration walk (snapshot_durations.go), so a fifth home for a token
// is covered the day it is added rather than the day someone remembers
// this file.

var (
	riderType      = reflect.TypeOf(ManaSpendRider{})
	riderReachOnce sync.Once
	riderReach     map[reflect.Type]bool
)

// reachesRider reports whether a value of type t can hold a
// ManaSpendRider, as a fixed point over every type reachable from
// GameSnapshot (buildDurationReach's method).
func reachesRider(t reflect.Type) bool {
	riderReachOnce.Do(func() {
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
		reach := map[reflect.Type]bool{riderType: true}
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
		riderReach = reach
	})
	return riderReach[derefType(t)]
}

// unknownManaRiders lists every rider in the snapshot naming a kind, or
// a count key, this binary does not read. A trigger rider's key is NOT
// refused: an unknown trigger key has always been the weaker-than-
// printed "does not fire" (applyManaSpendRidersLocked), and the
// registry is the catalog's, which a rollback can shrink.
func (s *GameSnapshot) unknownManaRiders() []string {
	seen := map[string]bool{}
	walkRiderValues(reflect.ValueOf(s), "", func(where string, r ManaSpendRider) {
		if !KnownManaRiderKind(r.Kind) {
			seen[fmt.Sprintf("mana spend rider kind %q at %s", r.Kind, where)] = true
		}
		if r.Count != "" {
			if _, ok := ManaRiderCountFor(r.Count); !ok {
				seen[fmt.Sprintf("mana spend rider count %q at %s", r.Count, where)] = true
			}
		}
	})
	out := make([]string, 0, len(seen))
	for k := range seen {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func walkRiderValues(v reflect.Value, where string, visit func(string, ManaSpendRider)) {
	for v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}
	t := v.Type()
	if t == riderType {
		visit(where, v.Interface().(ManaSpendRider))
		return
	}
	if !reachesRider(t) {
		return
	}
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			walkRiderValues(v.Index(i), where+"[]", visit)
		}
	case reflect.Map:
		iter := v.MapRange()
		for iter.Next() {
			walkRiderValues(iter.Value(), where+"{}", visit)
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
			walkRiderValues(fv, child, visit)
		}
	}
}
