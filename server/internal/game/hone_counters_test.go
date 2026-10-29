package game

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

// hone_counters_test.go — CR 122.1j, #4500: a hone counter on an
// Equipment gives +1/+0 to the creature it is attached to, as a layer 7c
// effect (CR 613.4c). The engine owns the rule, so these tests use bare
// Equipment (no catalog entry) and read the creature's power through the
// one accessor the wire and combat use.

// honePower is a creature's power after layers and counters.
func honePower(t *testing.T, g *Game, id uuid.UUID) int {
	t.Helper()
	c, ok := battlefieldCardByID(g, id)
	if !ok {
		t.Fatalf("card %s not on battlefield", id)
	}
	p, _, _ := c.PowerToughness()
	return p
}

func honeAttach(t *testing.T, g *Game, equipment, host uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(equipment, TargetRef{Kind: TargetCard, ID: host}); err != nil {
			t.Fatalf("AttachForEffect: %v", err)
		}
	})
}

func honeBoard(t *testing.T) (g *Game, bear, sword uuid.UUID) {
	t.Helper()
	g = newActiveGame(t)
	me := g.Seats[0]
	bear = pushScopedTestCreature(g, me.ID, 2, 2)
	sword = pushAttachTestCard(g, me.ID, "Test Sword", "Artifact — Equipment")
	return g, bear, sword
}

func TestHoneCounterKindIsKnown(t *testing.T) {
	if CounterHone != "hone" {
		t.Fatalf("CounterHone = %q, an on-disk counter name, must stay %q", CounterHone, "hone")
	}
	found := false
	for _, k := range KnownCardCounters {
		if k == CounterHone {
			found = true
		}
	}
	if !found {
		t.Error("CounterHone is missing from KnownCardCounters")
	}
	if _, _, ok := ParsePTCounter(CounterHone); ok {
		t.Error("a hone counter is not a P/T counter on its own creature (CR 122.1j puts it on an Equipment)")
	}
}

// Hone counters on an Equipment that is not attached give nothing.
func TestHoneCountersOnAnUnattachedEquipmentGiveNothing(t *testing.T) {
	g, bear, sword := honeBoard(t)
	addCounterForTest(t, g, sword, CounterHone, 3)
	if got := honePower(t, g, bear); got != 2 {
		t.Errorf("power = %d, want 2: an unattached Equipment gives no bonus", got)
	}
}

// Attaching gives +N/+0 for N hone counters, and only power.
func TestHoneCountersGivePlusNZeroToTheEquippedCreature(t *testing.T) {
	g, bear, sword := honeBoard(t)
	addCounterForTest(t, g, sword, CounterHone, 3)
	honeAttach(t, g, sword, bear)
	if got := honePower(t, g, bear); got != 5 {
		t.Errorf("power = %d, want 5 (2 + 3 hone counters)", got)
	}
	c, _ := battlefieldCardByID(g, bear)
	if _, _, toughness := c.PowerToughness(); toughness != 2 {
		t.Errorf("toughness = %d, want 2: hone is +1/+0", toughness)
	}
}

// The bonus tracks the counter count as counters are added and removed.
func TestHoneBonusTracksCounterChanges(t *testing.T) {
	g, bear, sword := honeBoard(t)
	honeAttach(t, g, sword, bear)
	if got := honePower(t, g, bear); got != 2 {
		t.Fatalf("power = %d before any counter, want 2", got)
	}
	addCounterForTest(t, g, sword, CounterHone, 1)
	if got := honePower(t, g, bear); got != 3 {
		t.Errorf("power = %d after one counter, want 3", got)
	}
	addCounterForTest(t, g, sword, CounterHone, 2)
	if got := honePower(t, g, bear); got != 5 {
		t.Errorf("power = %d after three counters, want 5", got)
	}
	addCounterForTest(t, g, sword, CounterHone, -2)
	if got := honePower(t, g, bear); got != 3 {
		t.Errorf("power = %d after removing two, want 3", got)
	}
	addCounterForTest(t, g, sword, CounterHone, -1)
	if got := honePower(t, g, bear); got != 2 {
		t.Errorf("power = %d after removing the last, want 2", got)
	}
}

// Re-equipping moves the bonus; so does unattaching it.
func TestHoneBonusFollowsTheEquipment(t *testing.T) {
	g, bear, sword := honeBoard(t)
	other := pushScopedTestCreature(g, g.Seats[0].ID, 1, 1)
	addCounterForTest(t, g, sword, CounterHone, 2)

	honeAttach(t, g, sword, bear)
	if honePower(t, g, bear) != 4 || honePower(t, g, other) != 1 {
		t.Fatalf("attached to the first bear: %d and %d, want 4 and 1", honePower(t, g, bear), honePower(t, g, other))
	}
	honeAttach(t, g, sword, other)
	if honePower(t, g, bear) != 2 || honePower(t, g, other) != 3 {
		t.Errorf("moved to the second bear: %d and %d, want 2 and 3", honePower(t, g, bear), honePower(t, g, other))
	}
	g.WithWriteLock(func() {
		if err := g.UnattachForEffect(sword); err != nil {
			t.Fatalf("UnattachForEffect: %v", err)
		}
	})
	if honePower(t, g, bear) != 2 || honePower(t, g, other) != 1 {
		t.Errorf("unattached: %d and %d, want 2 and 1", honePower(t, g, bear), honePower(t, g, other))
	}
}

// Two Equipment on one creature each contribute their own counters.
func TestHoneBonusesFromTwoEquipmentStack(t *testing.T) {
	g, bear, sword := honeBoard(t)
	axe := pushAttachTestCard(g, g.Seats[0].ID, "Test Axe", "Artifact — Equipment")
	addCounterForTest(t, g, sword, CounterHone, 2)
	addCounterForTest(t, g, axe, CounterHone, 3)
	honeAttach(t, g, sword, bear)
	if got := honePower(t, g, bear); got != 4 {
		t.Fatalf("power = %d with one Equipment, want 4", got)
	}
	honeAttach(t, g, axe, bear)
	if got := honePower(t, g, bear); got != 7 {
		t.Errorf("power = %d with both Equipment, want 7 (2 + 2 + 3)", got)
	}
}

// The bonus stacks with the creature's own counters, and it is a layer
// 7c modify: a 7b "base power 1" is applied first whatever the order
// the two began in (CR 613.4b, 613.4c).
func TestHoneBonusIsLayer7cAfterASetEffect(t *testing.T) {
	steppedClock(t)
	g, bear, sword := honeBoard(t)
	addCounterForTest(t, g, sword, CounterHone, 2)
	honeAttach(t, g, sword, bear)
	addCounterForTest(t, g, bear, CounterPlusOne, 1)
	if got := honePower(t, g, bear); got != 5 {
		t.Fatalf("power = %d, want 5 (2 + 2 hone + one +1/+1 counter)", got)
	}
	// The set effect is registered AFTER the attach, so it is the newer
	// timestamp; a 7b set still comes before the 7c modify.
	registerScopedEffectForTest(t, g, bear, []Mod{SetBasePowerMod(1)}, IndefiniteDuration())
	if got := honePower(t, g, bear); got != 4 {
		t.Errorf("power = %d, want 4 (base 1 + 2 hone + one +1/+1 counter)", got)
	}
}

// A hone counter on something that is not an Equipment is storage.
func TestHoneCountersOnANonEquipmentGiveNothing(t *testing.T) {
	g, bear, _ := honeBoard(t)
	aura := pushAttachTestCard(g, g.Seats[0].ID, "Test Aura", "Enchantment — Aura")
	addCounterForTest(t, g, aura, CounterHone, 4)
	honeAttach(t, g, aura, bear)
	if got := honePower(t, g, bear); got != 2 {
		t.Errorf("power = %d, want 2: only an Equipment's hone counters count (CR 122.1j)", got)
	}
}

// An Equipment's controller need not be the creature's: the bonus goes to
// whatever the Equipment is attached to.
func TestHoneBonusGoesToACreatureTheEquipmentsControllerDoesNotControl(t *testing.T) {
	g, _, sword := honeBoard(t)
	theirs := pushScopedTestCreature(g, g.Seats[1].ID, 3, 3)
	addCounterForTest(t, g, sword, CounterHone, 1)
	honeAttach(t, g, sword, theirs)
	if got := honePower(t, g, theirs); got != 4 {
		t.Errorf("power = %d, want 4", got)
	}
}

// An Equipment that leaves the battlefield takes its bonus with it.
func TestHoneBonusEndsWhenTheEquipmentLeaves(t *testing.T) {
	g, bear, sword := honeBoard(t)
	addCounterForTest(t, g, sword, CounterHone, 2)
	honeAttach(t, g, sword, bear)
	if got := honePower(t, g, bear); got != 4 {
		t.Fatalf("power = %d, want 4", got)
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(sword); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
	if got := honePower(t, g, bear); got != 2 {
		t.Errorf("power = %d after the Equipment was destroyed, want 2", got)
	}
}

// Nothing new is stored: the counter and the attachment are all a
// snapshot needs, and the bonus is there again after a restore. A
// snapshot from before hone counters existed has no such key, and reads
// as no bonus.
func TestHoneBonusSurvivesASnapshotRoundTrip(t *testing.T) {
	g, bear, sword := honeBoard(t)
	addCounterForTest(t, g, sword, CounterHone, 2)
	honeAttach(t, g, sword, bear)

	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatalf("marshal snapshot: %v", err)
	}
	var snap GameSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatalf("decode snapshot: %v", err)
	}
	restored, err := snap.Restore()
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	if got := honePower(t, restored, bear); got != 4 {
		t.Errorf("restored power = %d, want 4", got)
	}
	c, ok := battlefieldCardByID(restored, sword)
	if !ok || c.Counters[CounterHone] != 2 {
		t.Errorf("restored Equipment counters = %v, want hone 2", c.Counters)
	}
}
