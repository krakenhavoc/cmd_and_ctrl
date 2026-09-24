package game

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// snapshot_scoped_effects.go is the snapshot half of ADR 0041 phase 3's
// data records (#1497): the deep copy at the snapshot boundary, and the
// restore-time refusal of an effect key this binary cannot interpret
// (ErrUnknownEffectKey, ADR 0041 P4). The record itself, its vocabulary
// and its interpreter are in scoped_effects.go.

// ---------------------------------------------------------------
// Restore validation
// ---------------------------------------------------------------

// checkEffectKeys refuses a snapshot naming an effect key this binary
// cannot interpret (ADR 0041 P4): a mod kind, or a delayed-trigger or
// stack-item body or condition key. Every key a binary can write is
// one it registered, so the only way to meet an unknown one is a
// ROLLBACK — the file was written by a newer build — and the answer is
// the one Decision 5 gives a too-new file: refuse it, keep it.
//
// Effect-body and condition keys are ADR 0041 P2's (tier 2): a key
// this binary registered restores; any other is refused.
func (s *GameSnapshot) checkEffectKeys() error {
	unknown := append([]string(nil), s.unknownEffectFields...)
	for _, e := range s.ScopedEffects {
		for _, m := range e.Mods {
			if !KnownModKind(m.Kind) {
				unknown = append(unknown, "mod kind "+string(m.Kind))
			}
		}
		if !e.Duration.Known() {
			unknown = append(unknown, fmt.Sprintf("scoped-effect duration kind %d / condition %d",
				e.Duration.Kind, e.Duration.Condition))
		}
	}
	for _, d := range s.DelayedTriggers {
		if d.Duration != nil && !d.Duration.Known() {
			unknown = append(unknown, fmt.Sprintf("delayed-trigger duration kind %d / condition %d",
				d.Duration.Kind, d.Duration.Condition))
		}
		if d.Body != "" && !KnownEffectBody(d.Body) {
			unknown = append(unknown, "delayed-trigger body "+d.Body)
		}
		if d.Condition != "" && !KnownEffectCondition(d.Condition) {
			unknown = append(unknown, "delayed-trigger condition "+d.Condition)
		}
		// A spell filter is a closed vocabulary too (#1568 review).
		for _, p := range []*EffectParams{d.Params, d.CondParams} {
			if p != nil && !p.Filter.Valid() {
				unknown = append(unknown, fmt.Sprintf("delayed-trigger spell filter %v", p.Filter.Types))
			}
		}
	}
	for _, list := range [][]stackItemSnapshot{s.StackMeta, s.PendingTriggers} {
		for _, it := range list {
			if it.Body != "" && !KnownEffectBody(it.Body) {
				unknown = append(unknown, "stack-item body "+it.Body)
			}
			if it.Params != nil && !it.Params.Filter.Valid() {
				unknown = append(unknown, fmt.Sprintf("stack-item spell filter %v", it.Params.Filter.Types))
			}
		}
	}
	if len(unknown) == 0 {
		return nil
	}
	return fmt.Errorf("%w: %s", ErrUnknownEffectKey, strings.Join(unknown, ", "))
}

// deepCopyScopedEffects copies records and their inner slices, for the
// snapshot boundary, where nothing may alias live game state.
func deepCopyScopedEffects(in []ScopedEffect) []ScopedEffect {
	if len(in) == 0 {
		return nil
	}
	out := make([]ScopedEffect, len(in))
	for i, e := range in {
		e.Affected = append([]AffectedObject(nil), e.Affected...)
		e.Mods = cloneMods(e.Mods)
		out[i] = e
	}
	return out
}

// unknownScopedEffectFields re-reads the raw snapshot and lists every
// key this binary's effect types do not declare. encoding/json drops an
// unknown key without a word, and on these types an unknown key is
// exactly what a newer build's vocabulary looks like: a field a future
// mod kind reads, a duration field a future condition needs, a param a
// future body reads. So it is refused (ErrUnknownEffectKey), never
// dropped. ADR 0041 P4.
//
// Covered: a scopedEffects record, its affected members, its mods and
// its duration (tier 1); and — #1568 review — every EffectParams a
// delayed trigger or a stack item carries (`params`, `condParams`),
// down through its `filter` and its `object`.
func unknownScopedEffectFields(data []byte) ([]string, error) {
	var envelope struct {
		ScopedEffects   []map[string]json.RawMessage `json:"scopedEffects"`
		DelayedTriggers []map[string]json.RawMessage `json:"delayedTriggers"`
		StackMeta       []map[string]json.RawMessage `json:"stackMeta"`
		PendingTriggers []map[string]json.RawMessage `json:"pendingTriggers"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, err
	}
	var out []string
	check := func(where string, obj map[string]json.RawMessage, known map[string]bool) {
		for k := range obj {
			if !known[k] {
				out = append(out, fmt.Sprintf("unknown field %q on %s", k, where))
			}
		}
	}
	object := func(raw json.RawMessage) (map[string]json.RawMessage, error) {
		if len(raw) == 0 || string(raw) == "null" {
			return nil, nil
		}
		var obj map[string]json.RawMessage
		err := json.Unmarshal(raw, &obj)
		return obj, err
	}
	nested := func(where string, raw json.RawMessage, known map[string]bool, many bool) error {
		if len(raw) == 0 || string(raw) == "null" {
			return nil
		}
		if !many {
			obj, err := object(raw)
			if err != nil {
				return err
			}
			check(where, obj, known)
			return nil
		}
		var objs []map[string]json.RawMessage
		if err := json.Unmarshal(raw, &objs); err != nil {
			return err
		}
		for _, obj := range objs {
			check(where, obj, known)
		}
		return nil
	}
	params := func(where string, raw json.RawMessage) error {
		obj, err := object(raw)
		if err != nil || obj == nil {
			return err
		}
		check(where, obj, effectParamsJSONKeys)
		if err := nested(where+" filter", obj["filter"], castFilterJSONKeys, false); err != nil {
			return err
		}
		return nested(where+" object", obj["object"], objectRefJSONKeys, false)
	}
	for _, rec := range envelope.ScopedEffects {
		check("a scoped effect", rec, scopedEffectJSONKeys)
		if err := nested("an affected object", rec["affected"], affectedObjectJSONKeys, true); err != nil {
			return nil, err
		}
		if err := nested("a mod", rec["mods"], modJSONKeys, true); err != nil {
			return nil, err
		}
		if err := nested("a duration", rec["duration"], durationJSONKeys, false); err != nil {
			return nil, err
		}
	}
	for _, dt := range envelope.DelayedTriggers {
		if err := params("a delayed trigger's params", dt["params"]); err != nil {
			return nil, err
		}
		if err := params("a delayed trigger's condParams", dt["condParams"]); err != nil {
			return nil, err
		}
	}
	for _, list := range [][]map[string]json.RawMessage{envelope.StackMeta, envelope.PendingTriggers} {
		for _, it := range list {
			if err := params("a stack item's params", it["params"]); err != nil {
				return nil, err
			}
		}
	}
	sort.Strings(out)
	return out, nil
}

var (
	scopedEffectJSONKeys   = jsonKeysOf(reflect.TypeOf(ScopedEffect{}))
	affectedObjectJSONKeys = jsonKeysOf(reflect.TypeOf(AffectedObject{}))
	modJSONKeys            = jsonKeysOf(reflect.TypeOf(Mod{}))
	durationJSONKeys       = jsonKeysOf(reflect.TypeOf(Duration{}))
	effectParamsJSONKeys   = jsonKeysOf(reflect.TypeOf(EffectParams{}))
	castFilterJSONKeys     = jsonKeysOf(reflect.TypeOf(CastFilter{}))
	objectRefJSONKeys      = jsonKeysOf(reflect.TypeOf(ObjectRef{}))
)

// jsonKeysOf is the set of keys encoding/json writes for a struct's
// exported fields: the tag name, or the field name when untagged.
func jsonKeysOf(t reflect.Type) map[string]bool {
	out := map[string]bool{}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.PkgPath != "" {
			continue
		}
		name := f.Name
		if tag, ok := f.Tag.Lookup("json"); ok {
			if n := strings.Split(tag, ",")[0]; n == "-" {
				continue
			} else if n != "" {
				name = n
			}
		}
		out[name] = true
	}
	return out
}
