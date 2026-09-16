package game

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

// combat_step_snapshot_test.go — the persistence half of #187 (ADR 0053
// Decision 1). Event.CombatStep and DamageAssignmentFrame.CombatStep are
// both new fields on types GameSnapshot embeds BY VALUE, so they are
// carried without a snapshot_drift_test.go classification entry
// (TestEmbeddedDomainTypesStayPureData keeps that shortcut honest) and
// TestSnapshotRoundTripIsExact covers the round trip once enrich sets
// them. What that property cannot see is a file written BEFORE the
// fields existed, which is what this test is for: no
// SnapshotSchemaVersion bump is needed only because such a file decodes
// to "" — untagged — and "" is the right answer for it.

// TestSnapshotWithoutCombatStepRestoresUntagged builds a pre-field file
// by taking an enriched snapshot and deleting the two keys (the fields
// are additive, so that is exactly what the old encoder wrote), then
// restores it.
func TestSnapshotWithoutCombatStepRestoresUntagged(t *testing.T) {
	g := newRestorableGame(t)
	enrich(t, g)

	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatalf("marshal snapshot: %v", err)
	}
	for _, key := range []string{`,"combat_step":"first_strike"`, `"CombatStep":"regular",`} {
		if n := bytes.Count(raw, []byte(key)); n != 1 {
			t.Fatalf("enriched snapshot has %d occurrences of %s, want 1 — enrich no longer tags what this test strips", n, key)
		}
		raw = bytes.Replace(raw, []byte(key), nil, 1)
	}
	if bytes.Contains(raw, []byte("combat_step")) || bytes.Contains(raw, []byte("CombatStep")) {
		t.Fatal("stripped snapshot still carries a combat step key; it is not a pre-field file")
	}

	var old GameSnapshot
	if err := json.Unmarshal(raw, &old); err != nil {
		t.Fatalf("decode pre-field snapshot: %v", err)
	}
	restored, err := old.Restore()
	if err != nil {
		t.Fatalf("restore pre-field snapshot: %v", err)
	}

	var sawCombat, sawFrame bool
	for _, ev := range restored.Events {
		if ev.Kind == EventDealDamage && ev.Combat {
			sawCombat = true
			if ev.CombatStep != "" {
				t.Errorf("pre-field combat damage event restored with CombatStep %q, want untagged", ev.CombatStep)
			}
		}
	}
	for _, pc := range restored.PendingChoices {
		if pc.DamageAssignment != nil {
			sawFrame = true
			if pc.DamageAssignment.CombatStep != "" {
				t.Errorf("pre-field damage-assignment frame restored with CombatStep %q, want untagged",
					pc.DamageAssignment.CombatStep)
			}
		}
	}
	if !sawCombat || !sawFrame {
		t.Fatalf("restored game lost the fixture: combat damage event %v, damage-assignment frame %v", sawCombat, sawFrame)
	}

	// Everything else in the old file survives unchanged: re-capturing
	// the restored game gives the file back, modulo the two fields the
	// round-trip test already normalises.
	after := restored.CaptureSnapshot()
	var want GameSnapshot
	if err := json.Unmarshal(raw, &want); err != nil {
		t.Fatalf("re-decode pre-field snapshot: %v", err)
	}
	want.TakenAt, after.TakenAt = time.Time{}, time.Time{}
	after.LayerVersion = want.LayerVersion
	if !reflect.DeepEqual(&want, after) {
		t.Errorf("pre-field snapshot did not restore faithfully.\nfile:     %s\nrestored: %s",
			mustJSON(t, &want), mustJSON(t, after))
	}
}
