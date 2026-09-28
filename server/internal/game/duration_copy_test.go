package game

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
)

// duration_copy_test.go is the engine half of #1593: a copy effect
// with a duration and a timestamp of its own, on a permanent already on
// the battlefield (CR 707.2, 611.2, 613.1a, 613.7). The catalog cards
// that drive it — and the ability-switching half, which needs real
// catalog entries — are tested in cards/effects/duration_copy_test.go.

// becomeCopyUntilEOT copies `of` onto `target` until end of turn, the
// way a resolving Mirage Mirror activation does.
func becomeCopyUntilEOT(t *testing.T, g *Game, target, of uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		v, ok := g.CopiableValuesForEffect(of)
		if !ok {
			t.Fatalf("no copiable values for %s", of)
		}
		if !g.BecomeCopyForEffect(target, of, []uuid.UUID{target}, v, g.UntilEndOfTurnDuration(), "probe copy") {
			t.Fatal("BecomeCopyForEffect registered nothing")
		}
	})
}

// cleanupSweep runs the CR 514.2 sweep the cleanup step runs.
func cleanupSweep(g *Game) {
	g.WithWriteLock(func() { g.ClearEndOfTurnScopedStaticsLocked() })
}

func copyProbe(t *testing.T, g *Game, id uuid.UUID) Card {
	t.Helper()
	var out Card
	found := false
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		if c, ok := g.battlefieldCardLocked(id); ok {
			out, found = *c, true
		}
	})
	if !found {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return out
}

func vanillaOnBattlefield(g *Game, owner uuid.UUID, name, typeLine string, p, tough int) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   typeLine,
		OracleID:   "oracle-" + name,
		ScryfallID: "scry-" + name,
		ManaCost:   "{3}",
		Power:      p,
		Toughness:  tough,
		Owner:      owner,
		Controller: owner,
	})
	g.mu.Lock()
	g.EmitEvent(Event{Kind: EventZoneMove, CardID: id, OldZone: ZoneHand, NewZone: ZoneBattlefield})
	g.mu.Unlock()
	return id
}

// TestDurationCopyAppliesThenRevertsAtCleanup is the whole effect in
// one: the copy lands on a permanent already in play, the layer pass
// reads its values, and the cleanup step's sweep puts the permanent's
// own values back — visibly, on the flat fields, without waiting for a
// recompute.
func TestDurationCopyAppliesThenRevertsAtCleanup(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	mirror := vanillaOnBattlefield(g, me.ID, "Mirage Mirror", "Artifact", 0, 0)
	bears := bearsOnBattlefield(g, me.ID, "Grizzly Bears")

	becomeCopyUntilEOT(t, g, mirror, bears)
	c := copyProbe(t, g, mirror)
	if c.Name != "Grizzly Bears" || c.OracleID != "oracle-Grizzly Bears" || !c.IsCreature() {
		t.Fatalf("copy did not land: name %q oracle %q creature %v", c.Name, c.OracleID, c.IsCreature())
	}
	if p, tough := c.Effective().Power, c.Effective().Toughness; p != 2 || tough != 2 {
		t.Errorf("copy P/T = %d/%d, want 2/2", p, tough)
	}
	if !c.IsCopy() || c.DurationCopyBase == nil || c.DurationCopyBase.Name != "Mirage Mirror" {
		t.Errorf("baseline not stashed: IsCopy %v base %+v", c.IsCopy(), c.DurationCopyBase)
	}

	cleanupSweep(g)
	var raw Card
	g.WithWriteLock(func() {
		r, _ := g.battlefieldCardLocked(mirror)
		raw = *r
	})
	if raw.Name != "Mirage Mirror" || raw.OracleID != "oracle-Mirage Mirror" {
		t.Fatalf("after cleanup the flat fields are %q / %q, want Mirage Mirror's own", raw.Name, raw.OracleID)
	}
	if raw.IsCopy() || raw.DurationCopyBase != nil {
		t.Errorf("after cleanup the permanent still reports a copy: PrintedSelf %+v base %+v", raw.PrintedSelf, raw.DurationCopyBase)
	}
	if c := copyProbe(t, g, mirror); c.IsCreature() {
		t.Error("after cleanup Mirage Mirror is still a creature")
	}
	if len(g.ScopedEffects) != 0 {
		t.Errorf("the record outlived its duration: %d left", len(g.ScopedEffects))
	}
}

// TestTwoDurationCopiesApplyInTimestampOrder — CR 613.7: the later
// timestamp wins, whatever order the records were filed in, and when
// it ends the earlier one is what shows.
func TestTwoDurationCopiesApplyInTimestampOrder(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	target := vanillaOnBattlefield(g, me.ID, "Shapeshifter", "Creature — Shapeshifter", 0, 1)
	older := bearsOnBattlefield(g, me.ID, "Grizzly Bears")
	newer := vanillaOnBattlefield(g, me.ID, "Hill Giant", "Creature — Giant", 3, 3)

	g.WithWriteLock(func() {
		ov, _ := g.CopiableValuesForEffect(older)
		nv, _ := g.CopiableValuesForEffect(newer)
		affected := g.PinnedObjectsLocked(target)
		// The NEWER timestamp is filed FIRST, so registration order
		// and timestamp order disagree.
		g.appendScopedEffectLocked(target, affected, ScopeNone, uuid.Nil, []Mod{BecomeCopyMod(nv)}, g.UntilEndOfTurnDuration(), "newer", 200)
		g.appendScopedEffectLocked(target, affected, ScopeNone, uuid.Nil, []Mod{BecomeCopyMod(ov)}, IndefiniteDuration(), "older", 100)
		g.materialiseDurationCopiesLocked()
	})
	if c := copyProbe(t, g, target); c.Name != "Hill Giant" || c.Effective().Power != 3 {
		t.Fatalf("with both copies live the permanent is %q %d/%d, want the later timestamp's Hill Giant",
			c.Name, c.Effective().Power, c.Effective().Toughness)
	}
	cleanupSweep(g)
	if c := copyProbe(t, g, target); c.Name != "Grizzly Bears" || c.Effective().Power != 2 {
		t.Fatalf("after the later copy ended the permanent is %q, want the earlier copy's Grizzly Bears", c.Name)
	}
}

// TestADurationCopyOnACloneRevertsToTheClonesCopy — the case the
// second baseline exists for. A Clone that entered as Grizzly Bears
// and was then Cytoshaped into a Hill Giant is Grizzly Bears again at
// cleanup, not Clone; and when it later dies it is Clone in the
// graveyard (CR 400.7).
func TestADurationCopyOnACloneRevertsToTheClonesCopy(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bears := bearsOnBattlefield(g, me.ID, "Grizzly Bears")
	giant := vanillaOnBattlefield(g, me.ID, "Hill Giant", "Creature — Giant", 3, 3)
	clone := vanillaOnBattlefield(g, me.ID, "Clone", "Creature — Shapeshifter", 0, 0)
	g.WithWriteLock(func() {
		src, _ := g.battlefieldCardLocked(bears)
		cl, _ := g.battlefieldCardLocked(clone)
		cl.applyCopy(CopiableValuesOf(*src), *src)
	})

	becomeCopyUntilEOT(t, g, clone, giant)
	if c := copyProbe(t, g, clone); c.Name != "Hill Giant" {
		t.Fatalf("Cytoshaped Clone is %q, want Hill Giant", c.Name)
	}
	cleanupSweep(g)
	c := copyProbe(t, g, clone)
	if c.Name != "Grizzly Bears" || c.OracleID != "oracle-Grizzly Bears" {
		t.Fatalf("after cleanup the Clone is %q (%s), want its entry copy, Grizzly Bears", c.Name, c.OracleID)
	}
	if !c.IsCopy() || c.PrintedSelf == nil || c.PrintedSelf.Name != "Clone" {
		t.Errorf("the Clone lost its own printed values: %+v", c.PrintedSelf)
	}
	if err := g.MoveCardByID(ZoneRef{Kind: ZoneBattlefield}, ZoneRef{Kind: ZoneGraveyard, Owner: me.ID}, clone); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	for _, gc := range me.Graveyard.Cards {
		if gc.InstanceID == clone && (gc.Name != "Clone" || gc.IsCopy() || gc.DurationCopyBase != nil) {
			t.Errorf("the dead Clone is %q, copy %v, base %+v — want the card itself", gc.Name, gc.IsCopy(), gc.DurationCopyBase)
		}
	}
}

// TestADurationCopyKeepsCountersAndDamage — CR 707.2: counters, damage
// and status are not copiable values, so the copy neither brings the
// copied creature's nor wipes the permanent's own.
func TestADurationCopyKeepsCountersAndDamage(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	target := vanillaOnBattlefield(g, me.ID, "Shapeshifter", "Creature — Shapeshifter", 1, 4)
	giant := vanillaOnBattlefield(g, me.ID, "Hill Giant", "Creature — Giant", 3, 3)
	if err := g.AddCounter(target, "+1/+1", 2); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if err := g.AddCounter(giant, "+1/+1", 5); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	g.WithWriteLock(func() {
		c, _ := g.battlefieldCardLocked(target)
		c.DamageMarked = 1
		c.Tapped = true
	})

	becomeCopyUntilEOT(t, g, target, giant)
	c := copyProbe(t, g, target)
	if c.Counters["+1/+1"] != 2 || c.DamageMarked != 1 || !c.Tapped {
		t.Errorf("the copy disturbed per-permanent state: counters %v damage %d tapped %v", c.Counters, c.DamageMarked, c.Tapped)
	}
	if got := c.CurrentPower(); got != 5 {
		t.Errorf("power = %d, want 5 (the Giant's printed 3 plus the permanent's own two counters, not the Giant's five)", got)
	}
}

// TestADurationCopyOutlivesTheCardItCopied — Shifting Woodland and
// Lazav copy a card in a graveyard, which can then leave. The record
// holds values, not a reference, so the copy does not notice.
func TestADurationCopyOutlivesTheCardItCopied(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	land := vanillaOnBattlefield(g, me.ID, "Shifting Woodland", "Land", 0, 0)
	dead := uuid.New()
	me.Graveyard.PushTop(Card{
		InstanceID: dead, Name: "Craterhoof Behemoth", TypeLine: "Creature — Beast",
		OracleID: "oracle-hoof", Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID,
	})
	becomeCopyUntilEOT(t, g, land, dead)
	if err := g.MoveCardByID(ZoneRef{Kind: ZoneGraveyard, Owner: me.ID}, ZoneRef{Kind: ZoneExile}, dead); err != nil {
		t.Fatalf("exile the copied card: %v", err)
	}
	g.BumpLayerVersionForTest()
	c := copyProbe(t, g, land)
	if c.Name != "Craterhoof Behemoth" || c.Effective().Power != 5 {
		t.Fatalf("after the copied card left the graveyard the land is %q %d/%d, want the copy to stand",
			c.Name, c.Effective().Power, c.Effective().Toughness)
	}
}

// TestADurationCopyEndsWhenThePermanentLeaves — CR 400.7 again. The
// record is pinned to the object, so it is collected, and the card that
// arrives in the graveyard is itself.
func TestADurationCopyEndsWhenThePermanentLeaves(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	mirror := vanillaOnBattlefield(g, me.ID, "Mirage Mirror", "Artifact", 0, 0)
	bears := bearsOnBattlefield(g, me.ID, "Grizzly Bears")
	becomeCopyUntilEOT(t, g, mirror, bears)
	if err := g.MoveCardByID(ZoneRef{Kind: ZoneBattlefield}, ZoneRef{Kind: ZoneGraveyard, Owner: me.ID}, mirror); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	for _, gc := range me.Graveyard.Cards {
		if gc.InstanceID == mirror && (gc.Name != "Mirage Mirror" || gc.IsCopy() || gc.DurationCopyBase != nil) {
			t.Errorf("the dead copy is %q (copy %v)", gc.Name, gc.IsCopy())
		}
	}
	if len(g.ScopedEffects) != 0 {
		t.Errorf("the pinned record outlived its object: %d left", len(g.ScopedEffects))
	}
}

// TestAnIndefiniteCopyDropsTheCopiesItSupersedes — Unstable
// Shapeshifter becomes a copy of every creature that enters. Each copy
// makes the previous one invisible for as long as the permanent lasts,
// so the registry keeps one record, not one per creature.
func TestAnIndefiniteCopyDropsTheCopiesItSupersedes(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	target := vanillaOnBattlefield(g, me.ID, "Unstable Shapeshifter", "Creature — Shapeshifter", 0, 1)
	var last uuid.UUID
	for _, name := range []string{"Grizzly Bears", "Runeclaw Bear", "Hill Giant"} {
		last = vanillaOnBattlefield(g, me.ID, name, "Creature — Bear", 2, 2)
		g.WithWriteLock(func() {
			v, _ := g.CopiableValuesForEffect(last)
			g.BecomeCopyForEffect(target, last, []uuid.UUID{target}, v, IndefiniteDuration(), "shapeshift")
		})
	}
	if n := len(g.ScopedEffects); n != 1 {
		t.Fatalf("registry holds %d copy records, want 1", n)
	}
	if c := copyProbe(t, g, target); c.Name != "Hill Giant" {
		t.Errorf("the Shapeshifter is %q, want the last creature, Hill Giant", c.Name)
	}
	cleanupSweep(g)
	if c := copyProbe(t, g, target); c.Name != "Hill Giant" {
		t.Errorf("an indefinite copy ended at cleanup: %q", c.Name)
	}
}

// TestADurationCopyRoundTripsUndoAndSnapshot — the record and the
// baseline are both data. An undo taken mid-copy restores a copy that
// still reverts correctly, and so does a restore point written to disk.
func TestADurationCopyRoundTripsUndoAndSnapshot(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bears := bearsOnBattlefield(g, me.ID, "Grizzly Bears")
	giant := vanillaOnBattlefield(g, me.ID, "Hill Giant", "Creature — Giant", 3, 3)
	clone := vanillaOnBattlefield(g, me.ID, "Clone", "Creature — Shapeshifter", 0, 0)
	g.WithWriteLock(func() {
		src, _ := g.battlefieldCardLocked(bears)
		cl, _ := g.battlefieldCardLocked(clone)
		cl.applyCopy(CopiableValuesOf(*src), *src)
	})
	becomeCopyUntilEOT(t, g, clone, giant)

	undo := g.Clone()
	cleanupSweep(g)
	if c := copyProbe(t, g, clone); c.Name != "Grizzly Bears" {
		t.Fatalf("live game after cleanup: %q", c.Name)
	}
	g.RestoreFrom(undo)
	if c := copyProbe(t, g, clone); c.Name != "Hill Giant" {
		t.Fatalf("undo restored %q, want the Cytoshaped Hill Giant", c.Name)
	}

	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("a game holding a duration copy is not a restore point: %+v", snap.Continuations)
	}
	raw, err := json.Marshal(snap)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded GameSnapshot
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	restored, err := decoded.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	if c := copyProbe(t, restored, clone); c.Name != "Hill Giant" || c.Effective().Power != 3 {
		t.Fatalf("restored copy is %q", c.Name)
	}
	cleanupSweep(restored)
	c := copyProbe(t, restored, clone)
	if c.Name != "Grizzly Bears" || c.PrintedSelf == nil || c.PrintedSelf.Name != "Clone" {
		t.Fatalf("restored game reverted to %q (own %+v), want Grizzly Bears over Clone", c.Name, c.PrintedSelf)
	}
}

// TestRestoreRefusesAMalformedCopyMod — the copy mod's values and the
// bundles its except clause granted are keys like any other: a record
// this binary could not have written is refused, never guessed at.
func TestRestoreRefusesAMalformedCopyMod(t *testing.T) {
	for name, mod := range map[string]Mod{
		"no values":              {Kind: ModBecomeCopy},
		"unknown bundle":         {Kind: ModBecomeCopy, Copy: []PrintedValues{{Name: "X", GrantedAbilities: []string{"grant:no-such-bundle"}}}},
		"values on another kind": {Kind: ModModifyPT, Power: 1, Copy: []PrintedValues{{Name: "X"}}},
	} {
		s := &GameSnapshot{ScopedEffects: []ScopedEffect{{
			Affected: []AffectedObject{{ID: uuid.New()}},
			Mods:     []Mod{mod},
		}}}
		if err := s.checkEffectKeys(); !errors.Is(err, ErrUnknownEffectKey) {
			t.Errorf("%s: err = %v, want ErrUnknownEffectKey", name, err)
		}
	}
}

// TestRegistrationRefusesAMalformedCopyMod — the same line at the
// other end: a programming error, caught by the first test.
func TestRegistrationRefusesAMalformedCopyMod(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	target := vanillaOnBattlefield(g, me.ID, "Shapeshifter", "Creature — Shapeshifter", 0, 1)
	defer func() {
		if recover() == nil {
			t.Error("a becomeCopy mod with no values registered")
		}
	}()
	g.WithWriteLock(func() {
		g.RegisterScopedEffectForEffect(target, g.PinnedObjectsLocked(target), []Mod{{Kind: ModBecomeCopy}}, IndefiniteDuration(), "bad")
	})
}
