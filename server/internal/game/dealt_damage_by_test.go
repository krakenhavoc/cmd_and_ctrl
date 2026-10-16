package game

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// dealt_damage_by_test.go — #1946. TurnTally.DamagedObjects records
// which permanent objects each source dealt damage to this turn, and the
// rules built on it: Runesword's ScopeDealtDamageBy records and Bone
// Shaman's DealtDamageCantBeRegenerated static. The cards are pinned in
// cards/effects.

func liveObjectRef(t *testing.T, g *Game, id uuid.UUID) ObjectRef {
	t.Helper()
	var ref ObjectRef
	var ok bool
	g.WithWriteLock(func() { ref, ok = g.PermanentRefForEffect(id) })
	if !ok {
		t.Fatalf("no permanent object for %s", id)
	}
	return ref
}

func dealtDamageTo(g *Game, source, victim uuid.UUID) bool {
	var got bool
	g.WithWriteLock(func() { got = g.PermanentDealtDamageToPermanentThisTurn(source, victim) })
	return got
}

func TestDamagedObjectsRecordsCombatAndNoncombatDamage(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	shaman := pushScopedTestCreature(g, me, 3, 3)
	blocker := pushScopedTestCreature(g, opp, 2, 4)
	pinged := pushScopedTestCreature(g, opp, 2, 4)
	idle := pushScopedTestCreature(g, opp, 2, 4)
	emit(g, Event{Kind: EventDealDamage, Actor: me, Source: shaman, Target: blocker, Amount: 3, Combat: true})
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(shaman, pinged, 1); err != nil {
			t.Fatal(err)
		}
	})
	if !dealtDamageTo(g, shaman, blocker) {
		t.Error("combat damage to a creature is not recorded")
	}
	if !dealtDamageTo(g, shaman, pinged) {
		t.Error("noncombat damage to a creature is not recorded")
	}
	if dealtDamageTo(g, shaman, idle) {
		t.Error("a creature nothing damaged is recorded")
	}
	if dealtDamageTo(g, blocker, shaman) {
		t.Error("the record is per source")
	}
	var recs []ObjectDamageRecord
	g.ReadSnapshot(func() { recs = g.TurnTally.DamagedObjects })
	if len(recs) != 2 || !recs[0].Combat || recs[1].Combat {
		t.Errorf("records = %+v, want one combat and one noncombat", recs)
	}
}

// Damage prevented to 0 was never dealt (CR 615.1).
func TestDamagedObjectsIgnoresDamagePreventedToZero(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	shaman := pushScopedTestCreature(g, me, 3, 3)
	bear := pushScopedTestCreature(g, opp, 2, 4)
	g.WithWriteLock(func() {
		if !g.PreventNextDamageThisTurnForEffect(uuid.Nil, bear, 2, false, "shield") {
			t.Fatal("no shield")
		}
		if err := g.DealDamageToCreatureForEffect(shaman, bear, 2); err != nil {
			t.Fatal(err)
		}
	})
	if dealtDamageTo(g, shaman, bear) {
		t.Error("damage prevented to 0 is recorded as dealt")
	}
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(shaman, bear, 1); err != nil {
			t.Fatal(err)
		}
	})
	if !dealtDamageTo(g, shaman, bear) {
		t.Error("damage the spent shield let through is not recorded")
	}
}

// CR 400.7: a creature that left and came back is a new object that was
// never dealt the damage, and a source that did the same has damaged
// nothing.
func TestDamagedObjectsTreatsAFlickeredCreatureAsNew(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	shaman := pushScopedTestCreature(g, me, 3, 3)
	bear := pushScopedTestCreature(g, opp, 2, 4)
	ox := pushScopedTestCreature(g, opp, 2, 4)
	emit(g,
		Event{Kind: EventDealDamage, Actor: me, Source: shaman, Target: bear, Amount: 3, Combat: true},
		Event{Kind: EventDealDamage, Actor: me, Source: shaman, Target: ox, Amount: 3, Combat: true},
	)
	oldBear := liveObjectRef(t, g, bear)
	flicker := func(id uuid.UUID) uuid.UUID {
		var back uuid.UUID
		g.WithWriteLock(func() {
			if err := g.ExileCardForEffect(id); err != nil {
				t.Fatal(err)
			}
			var err error
			if back, err = g.ReturnFromExileToBattlefieldForEffect(id, uuid.Nil, false); err != nil {
				t.Fatal(err)
			}
		})
		return back
	}
	bear2 := flicker(bear)
	if dealtDamageTo(g, shaman, bear2) {
		t.Error("a flickered creature is a new object nothing damaged")
	}
	var old bool
	shamanRef := liveObjectRef(t, g, shaman)
	g.WithWriteLock(func() { old = g.ObjectDealtDamageToObjectThisTurn(shamanRef, oldBear) })
	if !old {
		t.Error("the record of the old object is gone")
	}
	shaman2 := flicker(shaman)
	if dealtDamageTo(g, shaman2, ox) {
		t.Error("a flickered source is a new object that damaged nothing")
	}
}

func TestDamagedObjectsOutliveTheVictimAndResetNextTurn(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	shaman := pushScopedTestCreature(g, me, 3, 3)
	bear := pushScopedTestCreature(g, opp, 2, 2)
	emit(g, Event{Kind: EventDealDamage, Actor: me, Source: shaman, Target: bear, Amount: 3, Combat: true})
	destroyForTest(t, g, bear)
	if !dealtDamageTo(g, shaman, bear) {
		t.Error("a creature that died is read as the object that died")
	}
	seat := g.Turn.ActiveSeat
	for i := 0; i < 40 && g.Turn.ActiveSeat == seat; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	var n int
	g.ReadSnapshot(func() { n = len(g.TurnTally.DamagedObjects) })
	if n != 0 {
		t.Errorf("last turn's damage is still on the record: %d entries", n)
	}
}

func TestDamagedObjectsSurviveCloneRestoreAndTheSnapshotWire(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	shaman := pushScopedTestCreature(g, me, 3, 3)
	bear := pushScopedTestCreature(g, opp, 2, 4)
	ox := pushScopedTestCreature(g, opp, 2, 4)
	emit(g, Event{Kind: EventDealDamage, Actor: me, Source: shaman, Target: bear, Amount: 3, Combat: true})

	snap := g.Clone()
	emit(g, Event{Kind: EventDealDamage, Actor: me, Source: shaman, Target: ox, Amount: 3, Combat: true})
	if dealtDamageTo(snap, shaman, ox) {
		t.Error("the clone shares its record with the live game")
	}
	g.WithWriteLock(func() { g.RestoreFrom(snap) })
	if !dealtDamageTo(g, shaman, bear) || dealtDamageTo(g, shaman, ox) {
		t.Error("restore did not bring back the record as it was")
	}

	raw, err := json.Marshal(g.TurnTally)
	if err != nil {
		t.Fatal(err)
	}
	var back TurnTally
	if err := json.Unmarshal(raw, &back); err != nil {
		t.Fatal(err)
	}
	if len(back.DamagedObjects) != 1 || back.DamagedObjects[0].Victim.ID != bear || back.DamagedObjects[0].Source.ID != shaman {
		t.Errorf("the wire lost the record: %s", raw)
	}
	var empty TurnTally
	raw, _ = json.Marshal(empty)
	if strings.Contains(string(raw), "damagedObjects") {
		t.Errorf("an empty tally must stay additive (omitempty): %s", raw)
	}
}

// Runesword's riders: a ScopeDealtDamageBy record reaches every creature
// its dealer damages, combat or not, and outlasts the dealer.
func TestDealtDamageByRecordsRefuseRegenerationAndExile(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	attacker := pushScopedTestCreature(g, me, 4, 4)
	early := pushScopedTestCreature(g, opp, 2, 2)
	late := pushScopedTestCreature(g, opp, 2, 2)
	bystander := pushScopedTestCreature(g, opp, 2, 2)
	emit(g, Event{Kind: EventDealDamage, Actor: me, Source: attacker, Target: early, Amount: 1, Combat: true})
	ref := liveObjectRef(t, g, attacker)
	g.WithWriteLock(func() {
		if !g.CantBeRegeneratedIfDealtDamageByForEffect(uuid.Nil, ref, true, "Runesword — no regeneration") {
			t.Fatal("no regeneration record")
		}
		if !g.ExileIfDealtDamageByWouldDieThisTurnForEffect(uuid.Nil, ref, me, false, "Runesword — exile") {
			t.Fatal("no exile record")
		}
	})
	emit(g, Event{Kind: EventDealDamage, Actor: me, Source: attacker, Target: late, Amount: 1, Combat: true})

	g.WithWriteLock(func() {
		if g.PermanentCantBeRegeneratedForEffect(early) {
			t.Error("damage dealt before \"if the creature deals damage\" resolved refuses regeneration")
		}
		if !g.PermanentCantBeRegeneratedForEffect(late) {
			t.Error("damage dealt after it resolved does not refuse regeneration")
		}
		if g.PermanentCantBeRegeneratedForEffect(bystander) {
			t.Error("a creature it never damaged can't be regenerated")
		}
		if got := g.ExileIfItWouldDieLabels(early); len(got) != 1 {
			t.Errorf("the exile chip on the earlier victim = %v", got)
		}
		if got := g.ExileIfCreaturesWouldDieThisTurnLabels(); len(got) != 0 {
			t.Errorf("Runesword's record is not a game-wide banner: %v", got)
		}
		if err := g.RegenerateForEffect(late); err != nil {
			t.Fatal(err)
		}
	})
	// The dealer leaves; the creatures it damaged are what the effect is
	// about, so it still applies.
	destroyForTest(t, g, attacker)
	destroyForTest(t, g, late)
	if z := zoneOf(t, g, late); z != ZoneExile {
		t.Errorf("the regeneration-refused victim went to %q, want exile", z)
	}
	// "Dealt damage by the targeted creature" is the whole turn's damage.
	destroyForTest(t, g, early)
	if z := zoneOf(t, g, early); z != ZoneExile {
		t.Errorf("the earlier victim went to %q, want exile", z)
	}
	destroyForTest(t, g, bystander)
	if z := zoneOf(t, g, bystander); z != ZoneGraveyard {
		t.Errorf("a creature the dealer never damaged went to %q, want the graveyard", z)
	}
}

// Bone Shaman's static is read live under the permanent's catalog key:
// it covers creatures damaged before it, and stops once the permanent
// has gone.
func TestDealtDamageCantBeRegeneratedStaticIsReadLive(t *testing.T) {
	const key = "dealt-damage-no-regen-fixture"
	prev := CatalogDealtDamageCantBeRegenerated
	CatalogDealtDamageCantBeRegenerated = func(k string) []DealtDamageCantBeRegenerated {
		if k == key {
			return []DealtDamageCantBeRegenerated{{Label: "no regen"}}
		}
		return nil
	}
	t.Cleanup(func() { CatalogDealtDamageCantBeRegenerated = prev })

	g := newActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	shaman := NewCard("Bone Shaman", me)
	shaman.TypeLine, shaman.Power, shaman.Toughness, shaman.OracleID = "Creature — Giant Shaman", 3, 3, key
	g.Battlefield.PushTop(shaman)
	bear := pushScopedTestCreature(g, opp, 2, 4)
	emit(g, Event{Kind: EventDealDamage, Actor: me, Source: shaman.InstanceID, Target: bear, Amount: 3, Combat: true})
	g.WithWriteLock(func() {
		if !g.PermanentCantBeRegeneratedForEffect(bear) {
			t.Error("a creature the static's permanent damaged can be regenerated")
		}
		if err := g.RegenerateForEffect(bear); err != nil {
			t.Fatal(err)
		}
	})
	destroyForTest(t, g, bear)
	if z := zoneOf(t, g, bear); z != ZoneGraveyard {
		t.Errorf("the refused shield regenerated it: %q", z)
	}
	if n := func() int {
		var n int
		g.ReadSnapshot(func() { n = g.RegenerationShieldsOn(bear) })
		return n
	}(); n != 0 {
		t.Errorf("shields left on a card in the graveyard: %d", n)
	}

	other := pushScopedTestCreature(g, opp, 2, 4)
	emit(g, Event{Kind: EventDealDamage, Actor: me, Source: shaman.InstanceID, Target: other, Amount: 1})
	removeFromBattlefieldForTest(t, g, shaman.InstanceID, true)
	g.WithWriteLock(func() {
		if g.PermanentCantBeRegeneratedForEffect(other) {
			t.Error("the static outlived its permanent")
		}
	})
}

func TestDealtDamageByScopeProblem(t *testing.T) {
	dealer := ObjectRef{ID: uuid.New(), Epoch: 1}
	for _, tc := range []struct {
		name  string
		scope AffectedScope
		mods  []Mod
		bad   bool
	}{
		{"sound", ScopeDealtDamageBy, []Mod{{Kind: ModCantBeRegenerated, Objects: []ObjectRef{dealer}, DamageSince: 2}}, false},
		{"no dealer", ScopeDealtDamageBy, []Mod{{Kind: ModExileIfWouldDie}}, true},
		{"other kind", ScopeDealtDamageBy, []Mod{{Kind: ModCantBeSacrificed, Objects: []ObjectRef{dealer}}}, true},
		{"negative since", ScopeDealtDamageBy, []Mod{{Kind: ModExileIfWouldDie, Objects: []ObjectRef{dealer}, DamageSince: -1}}, true},
		{"since elsewhere", ScopeCreatures, []Mod{{Kind: ModExileIfWouldDie, DamageSince: 1}}, true},
		{"dealer elsewhere", ScopeNone, []Mod{{Kind: ModCantBeRegenerated, Objects: []ObjectRef{dealer}}}, true},
		{"pinned record", ScopeNone, []Mod{{Kind: ModCantBeRegenerated}}, false},
	} {
		if got := dealtDamageByScopeProblem(tc.scope, tc.mods) != ""; got != tc.bad {
			t.Errorf("%s: problem = %v, want %v", tc.name, got, tc.bad)
		}
	}
}
