package game

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// #4361 S18: unknownScopedEffectFields held only the top-level
// scopedEffects list to this binary's keys. A record a keyed tail
// carries (KeyedTail.Records, #4361 S12) or an extra turn's rules
// effect (ExtraTurn.Effects, ADR 0188) is the same record, and a key a
// newer binary added to it was dropped by encoding/json without a word.

// TestScopedEffectRoutesCoverTailsAndExtraTurns: the routes are read off
// the snapshot types, so the tail's and the extra turn's records are
// found without being named.
func TestScopedEffectRoutesCoverTailsAndExtraTurns(t *testing.T) {
	routes := scopedEffectRoutes()
	for _, want := range []string{
		"scopedEffects[]",
		"extraTurns[].effects[]",
		"pendingChoices[].scryThen.records[]",
		"pendingChoices[].replacementResume.event.entryTail.keyed.records[]",
	} {
		if !slices.Contains(routes, want) {
			t.Errorf("route %s missing from %v", want, routes)
		}
	}
}

// storedRecordsSnapshotJSON is a snapshot of the current schema with a
// record on a pending choice's keyed tail and one on an extra turn,
// decoded to a generic map so a test can add a key a newer binary
// might write.
func storedRecordsSnapshotJSON(t *testing.T) map[string]any {
	t.Helper()
	rec := ScopedEffect{
		Seq:        7,
		SourceName: "records probe",
		Source:     ObjectRef{ID: uuid.NewSHA1(uuid.Nil, []byte("records probe")), Epoch: 1},
		Mods:       []Mod{{Kind: ModSetController}},
	}
	s := GameSnapshot{
		Schema: SnapshotSchemaVersion,
		PendingChoices: []pendingChoiceSnapshot{{
			ScryThen: &keyedTailSnapshot{Key: "probe/tail", Records: []ScopedEffect{rec},
				Next: &keyedTailSnapshot{Key: "probe/next", Records: []ScopedEffect{rec}}},
		}},
		ExtraTurns: []ExtraTurn{{Ref: 1, Seat: 1, Effects: []ScopedEffect{rec}}},
	}
	raw, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	return m
}

func firstObject(t *testing.T, v any, keys ...string) map[string]any {
	t.Helper()
	for _, k := range keys {
		obj, ok := v.(map[string]any)
		if !ok {
			t.Fatalf("%q: not an object: %v", k, v)
		}
		v = obj[k]
		if list, ok := v.([]any); ok {
			if len(list) == 0 {
				t.Fatalf("%q: empty list", k)
			}
			v = list[0]
		}
	}
	obj, ok := v.(map[string]any)
	if !ok {
		t.Fatalf("not an object: %v", v)
	}
	return obj
}

func TestUnknownKeyOnAStoredRecordIsRefused(t *testing.T) {
	base := storedRecordsSnapshotJSON(t)
	raw, err := json.Marshal(base)
	if err != nil {
		t.Fatal(err)
	}
	if got, err := unknownScopedEffectFields(raw, SnapshotSchemaVersion); err != nil || len(got) != 0 {
		t.Fatalf("a snapshot this binary wrote: unknown = %v, err = %v, want none", got, err)
	}

	cases := []struct {
		name string
		path []string
		want string
	}{
		{"a keyed tail's record", []string{"pendingChoices", "scryThen", "records"}, `on a scoped effect at pendingChoices[].scryThen.records[]`},
		{"a record on a tail's next tail", []string{"pendingChoices", "scryThen", "next", "records"}, `on a scoped effect at pendingChoices[].scryThen.next.records[]`},
		{"a mod of a keyed tail's record", []string{"pendingChoices", "scryThen", "records", "mods"}, `on a mod`},
		{"an extra turn's effect", []string{"extraTurns", "effects"}, `on a scoped effect at extraTurns[].effects[]`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := storedRecordsSnapshotJSON(t)
			firstObject(t, m, tc.path...)["fromANewerBinary"] = 1
			raw, err := json.Marshal(m)
			if err != nil {
				t.Fatal(err)
			}
			got, err := unknownScopedEffectFields(raw, SnapshotSchemaVersion)
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != 1 || !strings.Contains(got[0], `"fromANewerBinary"`) || !strings.Contains(got[0], tc.want) {
				t.Fatalf("unknown = %v, want one naming %q", got, tc.want)
			}
			// The decode path keeps it, so a restore refuses the file
			// (ErrUnknownEffectKey) rather than dropping the key.
			var s GameSnapshot
			if err := json.Unmarshal(raw, &s); err != nil {
				t.Fatal(err)
			}
			if len(s.unknownEffectFields) == 0 {
				t.Error("GameSnapshot.UnmarshalJSON dropped the unknown key")
			}
		})
	}
}
