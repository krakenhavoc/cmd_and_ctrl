package game

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

// combat_damage_steps_test.go — #717 and #716, the two halves of
// CR 510.4.
//
// #717: a combat in which any attacking or blocking creature has first
// strike or double strike as the combat damage step begins has TWO
// combat damage steps, and the active player receives priority after
// each (CR 510.3). The engine models that as a real step in
// turnSequence, `first_strike_damage`, that a turn only has when the
// condition holds — the same "a step this turn does not have is walked
// through" move a replacement-cancelled step makes.
//
// #716: which creatures deal damage in the SECOND step is decided as
// the FIRST one begins (CR 702.7c), not by re-reading keywords after
// the window between them. The window is where a lord dies and a pump
// lands, so the two readings genuinely differ.

const (
	fsLordOracle       = "test-first-strike-lord"
	dsLordOracle       = "test-double-strike-lord"
	fsDamageHookOracle = "test-first-strike-damage-hook"
)

// pushDamageStepCreature seeds a creature on the battlefield through
// the layer listener, so granted keywords and counters land on it the
// way they do in a real game. `keywords` are PRINTED keywords.
func pushDamageStepCreature(g *Game, owner *Player, name string, power, toughness int, keywords ...string) uuid.UUID {
	return pushDamageStepCard(g, Card{
		Name:       name,
		TypeLine:   "Creature — Test",
		Power:      power,
		Toughness:  toughness,
		Keywords:   append([]string(nil), keywords...),
		Owner:      owner.ID,
		Controller: owner.ID,
	})
}

// pushDamageStepCard is pushTypedTestCard plus "it has been here since
// the untap step": the entry event these fixtures need for the layer
// listener also marks the creature summoning-sick (CR 302.6), and
// every one of them is about to attack.
func pushDamageStepCard(g *Game, c Card) uuid.UUID {
	id := pushTypedTestCard(g, c)
	g.WithWriteLock(func() {
		if bc := findBattlefieldCard(g, id); bc != nil {
			bc.SummonedThisTurn = false
		}
	})
	return id
}

// keywordLordStatics is a Layer 6 grant of one keyword to the OTHER
// creatures its controller controls — Stromkirk Captain's shape, which
// is the shape #716 is about: the grant goes away with the lord.
func keywordLordStatics(keyword string) []StaticAbility {
	return []StaticAbility{{
		Layer: Layer6Ability,
		AppliesTo: func(target *Card, _ *Game, source *Card) bool {
			return target.InstanceID != source.InstanceID &&
				target.Controller == source.Controller &&
				target.IsCreature()
		},
		Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
			c.Abilities = append(c.Abilities, keyword)
		},
	}}
}

// declareAttacks walks the cursor to declare_attackers, declares each
// attacker against seat 1, and leaves the cursor in declare_blockers
// with nothing blocking unless the caller adds it.
func declareAttacks(t *testing.T, g *Game, attackers ...uuid.UUID) {
	t.Helper()
	advanceIntoStep(t, g, StepDeclareAttackers)
	for _, a := range attackers {
		if err := g.DeclareAttacker(a, g.Seats[1].ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	advanceIntoStep(t, g, StepDeclareBlockers)
}

// passUntilStep passes priority — the ordinary way a table moves —
// until the cursor reaches `want`. Every pass resolves whatever the
// previous step put on the stack, so this is the path a real game
// takes between the two combat damage steps.
func passUntilStep(t *testing.T, g *Game, want Step) {
	t.Helper()
	for range 32 {
		if g.Turn.Step == want {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority toward %s: %v", want, err)
		}
	}
	t.Fatalf("never reached %s by passing priority (at %s)", want, g.Turn.Step)
}

// stepBeganSeq returns the Seq of the one EventStepBegan for `step`,
// or 0 when the step never began. Fails when it began twice.
func stepBeganSeq(t *testing.T, g *Game, step Step) uint64 {
	t.Helper()
	var seq uint64
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == EventStepBegan && ev.Step == step {
			n++
			seq = ev.Seq
		}
	}
	if n > 1 {
		t.Fatalf("%s began %d times, want at most once", step, n)
	}
	return seq
}

// onBattlefield reports whether the card is still there.
func onBattlefield(g *Game, id uuid.UUID) bool {
	return findBattlefieldCard(g, id) != nil
}

// --- #717: the step exists, or it does not -------------------------

// A first striker in combat gives the turn two combat damage steps,
// each with its own EventStepBegan and its own priority window. The
// first striker's damage lands in the first; the vanilla attacker's in
// the second, and not before.
func TestTwoCombatDamageStepsWithAFirstStriker(t *testing.T) {
	g := newActiveGame(t)
	atk, def := g.Seats[0], g.Seats[1]
	striker := pushDamageStepCreature(g, atk, "Youthful Knight", 2, 1, "first strike")
	bear := pushDamageStepCreature(g, atk, "Grizzly Bears", 2, 2)
	declareAttacks(t, g, striker, bear)

	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep out of declare_blockers: %v", err)
	}
	if g.Turn.Step != StepFirstStrikeDamage {
		t.Fatalf("step after declare_blockers: got %q, want %q", g.Turn.Step, StepFirstStrikeDamage)
	}
	if g.Turn.Phase != PhaseCombat {
		t.Errorf("phase in the first damage step: got %q, want %q", g.Turn.Phase, PhaseCombat)
	}
	if g.Turn.PriorityHolder != g.Turn.ActiveSeat {
		t.Errorf("PriorityHolder in the first damage step: got %d, want the active seat %d",
			g.Turn.PriorityHolder, g.Turn.ActiveSeat)
	}
	if got := StartingLife - def.Life; got != 2 {
		t.Fatalf("defender lost %d in the first damage step, want 2 — first strike only", got)
	}

	fsSeq := stepBeganSeq(t, g, StepFirstStrikeDamage)
	if fsSeq == 0 {
		t.Fatal("no EventStepBegan for the first-strike damage step")
	}

	// The table passes priority, which is the window CR 510.3 asks
	// for, and lands in the second step.
	passUntilStep(t, g, StepCombatDamage)
	regSeq := stepBeganSeq(t, g, StepCombatDamage)
	if regSeq <= fsSeq {
		t.Fatalf("combat_damage began at seq %d, first_strike_damage at %d — want the first strike step first", regSeq, fsSeq)
	}
	if got := StartingLife - def.Life; got != 4 {
		t.Errorf("defender lost %d over both steps, want 4", got)
	}
	_ = bear
}

// With no first or double strike anywhere the turn has ONE combat
// damage step: the cursor never enters first_strike_damage, the step
// never announces, and declare_blockers is one advance away from
// combat damage — so no bot and no autopass sees an extra window.
func TestOneCombatDamageStepWithoutFirstStrike(t *testing.T) {
	g := newActiveGame(t)
	atk, def := g.Seats[0], g.Seats[1]
	bear := pushDamageStepCreature(g, atk, "Grizzly Bears", 2, 2)
	declareAttacks(t, g, bear)

	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep out of declare_blockers: %v", err)
	}
	if g.Turn.Step != StepCombatDamage {
		t.Fatalf("step after declare_blockers: got %q, want %q", g.Turn.Step, StepCombatDamage)
	}
	if seq := stepBeganSeq(t, g, StepFirstStrikeDamage); seq != 0 {
		t.Errorf("the first-strike damage step announced at seq %d, but nothing in combat has first strike", seq)
	}
	if got := StartingLife - def.Life; got != 2 {
		t.Errorf("defender lost %d, want 2", got)
	}
	// One priority pass around the table leaves combat damage, as it
	// always did.
	passUntilStep(t, g, StepEndCombat)
}

// A "whenever this deals combat damage" trigger from FIRST-STRIKE
// damage goes on the stack in the first step and resolves in the
// window, BEFORE the second step's damage is dealt. Here it destroys
// the vanilla attacker, so that attacker never deals its damage at
// all — which is only observable because the trigger resolved first.
func TestFirstStrikeDamageTriggerResolvesBeforeRegularDamage(t *testing.T) {
	g := newActiveGame(t)
	atk, def := g.Seats[0], g.Seats[1]
	var bear uuid.UUID
	withCatalogTriggers(t, func(oracle string) []TriggeredAbility {
		if oracle != fsDamageHookOracle {
			return nil
		}
		return []TriggeredAbility{{
			Watches: []EventKind{EventDealDamage},
			Key:     "first-strike damage hook",
			AppliesTo: func(ev Event, source *Card, _ Characteristic, _ *Game) bool {
				return ev.Combat && ev.Source == source.InstanceID
			},
			Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
				victim := bear
				return newTriggeredItemForTest(source, "destroy the other attacker", func(g *Game, _ *StackItem) error {
					return g.DestroyPermanentForEffect(victim)
				})
			},
		}}
	})

	striker := pushDamageStepCard(g, Card{
		Name: "Hooked Knight", TypeLine: "Creature — Test", OracleID: fsDamageHookOracle,
		Power: 2, Toughness: 1, Keywords: []string{"first strike"},
		Owner: atk.ID, Controller: atk.ID,
	})
	bear = pushDamageStepCreature(g, atk, "Grizzly Bears", 2, 2)
	declareAttacks(t, g, striker, bear)

	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep into the first damage step: %v", err)
	}
	if g.Turn.Step != StepFirstStrikeDamage {
		t.Fatalf("step: got %q, want %q", g.Turn.Step, StepFirstStrikeDamage)
	}
	if !onBattlefield(g, bear) {
		t.Fatal("the trigger resolved before anyone had priority")
	}

	passUntilStep(t, g, StepCombatDamage)
	if onBattlefield(g, bear) {
		t.Fatal("the first-strike damage trigger had not resolved by the second step")
	}
	if got := StartingLife - def.Life; got != 2 {
		t.Errorf("defender lost %d, want 2 — the Bears died in the window and never dealt damage", got)
	}
}

// The window is a real one: the board can change in it, and the second
// step reads the board it leaves behind. Two +1/+1 counters land on the
// vanilla attacker between the steps and its damage goes up.
func TestPumpInTheWindowChangesTheSecondStep(t *testing.T) {
	g := newActiveGame(t)
	atk, def := g.Seats[0], g.Seats[1]
	striker := pushDamageStepCreature(g, atk, "Youthful Knight", 2, 1, "first strike")
	bear := pushDamageStepCreature(g, atk, "Grizzly Bears", 2, 2)
	declareAttacks(t, g, striker, bear)

	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep into the first damage step: %v", err)
	}
	if err := g.AddCounter(bear, "+1/+1", 2); err != nil {
		t.Fatalf("AddCounter in the window: %v", err)
	}
	passUntilStep(t, g, StepCombatDamage)

	if got := StartingLife - def.Life; got != 6 {
		t.Errorf("defender lost %d, want 6 — 2 first strike then 4 regular", got)
	}
}

// Double strike deals in both steps (CR 702.4b).
func TestDoubleStrikeDealsInBothCombatDamageSteps(t *testing.T) {
	g := newActiveGame(t)
	atk, def := g.Seats[0], g.Seats[1]
	ace := pushDamageStepCreature(g, atk, "Fencing Ace", 2, 2, "double strike")
	declareAttacks(t, g, ace)

	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep into the first damage step: %v", err)
	}
	if got := StartingLife - def.Life; got != 2 {
		t.Fatalf("after the first step the defender lost %d, want 2", got)
	}
	passUntilStep(t, g, StepCombatDamage)
	if got := StartingLife - def.Life; got != 4 {
		t.Errorf("after both steps the defender lost %d, want 4", got)
	}
}

// --- #716: participation is fixed as the first step begins ---------

// The issue's own case. A Bear has first strike only because a lord
// grants it; the lord dies to first-strike damage in the first step, so
// by the second step the Bear reads as a non-first-striker. It must
// NOT deal damage a second time (CR 702.7c: the second step includes
// the creatures that had neither keyword as the FIRST step began).
func TestB716LordGrantingFirstStrikeDiesInTheFirstStep(t *testing.T) {
	g := newActiveGame(t)
	atk, def := g.Seats[0], g.Seats[1]
	withStaticAbilities(t, func(oracle string) []StaticAbility {
		if oracle != fsLordOracle {
			return nil
		}
		return keywordLordStatics("first strike")
	})

	lord := pushDamageStepCard(g, Card{
		Name: "Stromkirk Captain", TypeLine: "Creature — Test", OracleID: fsLordOracle,
		Power: 1, Toughness: 1, Owner: atk.ID, Controller: atk.ID,
	})
	bear := pushDamageStepCreature(g, atk, "Granted Striker", 2, 2)
	blocker := pushDamageStepCreature(g, def, "First Strike Blocker", 2, 2, "first strike")

	declareAttacks(t, g, lord, bear)
	if err := g.DeclareBlocker(blocker, lord); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}

	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep into the first damage step: %v", err)
	}
	if g.Turn.Step != StepFirstStrikeDamage {
		t.Fatalf("step: got %q, want %q", g.Turn.Step, StepFirstStrikeDamage)
	}
	if onBattlefield(g, lord) {
		t.Fatal("fixture: the lord should have died to the blocker's first-strike damage")
	}
	if got := StartingLife - def.Life; got != 2 {
		t.Fatalf("defender lost %d in the first step, want 2 — the granted striker only", got)
	}

	passUntilStep(t, g, StepCombatDamage)
	if got := StartingLife - def.Life; got != 2 {
		t.Errorf("defender lost %d over both steps, want 2 — the granted striker must not deal damage twice", got)
	}
}

// The mirror. A creature that GAINS first strike in the window between
// the steps had neither keyword when the first step began, so it deals
// its damage in the second step — it does not miss combat because the
// second step re-read its keywords and skipped it.
func TestB716CreatureGainingFirstStrikeInTheWindowStillDealsRegularDamage(t *testing.T) {
	g := newActiveGame(t)
	atk, def := g.Seats[0], g.Seats[1]
	withStaticAbilities(t, func(oracle string) []StaticAbility {
		if oracle != fsLordOracle {
			return nil
		}
		return keywordLordStatics("first strike")
	})

	striker := pushDamageStepCreature(g, atk, "Youthful Knight", 2, 1, "first strike")
	bear := pushDamageStepCreature(g, atk, "Grizzly Bears", 2, 2)
	declareAttacks(t, g, striker, bear)

	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep into the first damage step: %v", err)
	}
	if got := StartingLife - def.Life; got != 2 {
		t.Fatalf("defender lost %d in the first step, want 2", got)
	}
	// The lord arrives in the window and grants the Bears first strike
	// — too late to matter (CR 702.7c).
	pushTypedTestCard(g, Card{
		Name: "Stromkirk Captain", TypeLine: "Creature — Test", OracleID: fsLordOracle,
		Power: 1, Toughness: 1, Owner: atk.ID, Controller: atk.ID,
	})
	passUntilStep(t, g, StepCombatDamage)

	if got := StartingLife - def.Life; got != 4 {
		t.Errorf("defender lost %d over both steps, want 4 — the Bears deal regular damage", got)
	}
	if !HasKeyword(findBattlefieldCard(g, bear), "first strike") {
		t.Error("fixture: the lord should have granted the Bears first strike by the second step")
	}
}

// The double-strike half of CR 702.7c is the one clause the second step
// still reads LIVE: it includes the creatures that have double strike
// NOW. A creature whose granted double strike goes away in the window
// dealt its damage in the first step and deals none in the second.
func TestB716LosingDoubleStrikeInTheWindowStopsTheSecondHit(t *testing.T) {
	g := newActiveGame(t)
	atk, def := g.Seats[0], g.Seats[1]
	withStaticAbilities(t, func(oracle string) []StaticAbility {
		if oracle != dsLordOracle {
			return nil
		}
		return keywordLordStatics("double strike")
	})

	lord := pushDamageStepCard(g, Card{
		Name: "Double Strike Lord", TypeLine: "Creature — Test", OracleID: dsLordOracle,
		Power: 1, Toughness: 1, Owner: atk.ID, Controller: atk.ID,
	})
	bear := pushDamageStepCreature(g, atk, "Granted Double Striker", 2, 2)
	declareAttacks(t, g, bear)

	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep into the first damage step: %v", err)
	}
	if got := StartingLife - def.Life; got != 2 {
		t.Fatalf("defender lost %d in the first step, want 2", got)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(lord) })
	passUntilStep(t, g, StepCombatDamage)

	if got := StartingLife - def.Life; got != 2 {
		t.Errorf("defender lost %d over both steps, want 2 — the double strike went away in the window", got)
	}
}

// --- carried state: undo and snapshot inside the window ------------

// An undo that lands back in the window keeps the participation
// record, so replaying the second step does not let the first striker
// hit again.
func TestUndoIntoTheWindowKeepsTheParticipationRecord(t *testing.T) {
	g := newActiveGame(t)
	atk, def := g.Seats[0], g.Seats[1]
	striker := pushDamageStepCreature(g, atk, "Youthful Knight", 2, 1, "first strike")
	bear := pushDamageStepCreature(g, atk, "Grizzly Bears", 2, 2)
	declareAttacks(t, g, striker, bear)

	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep into the first damage step: %v", err)
	}
	inWindow := g.Clone()

	passUntilStep(t, g, StepCombatDamage)
	if got := StartingLife - def.Life; got != 4 {
		t.Fatalf("defender lost %d before the undo, want 4", got)
	}

	g.WithWriteLock(func() { g.RestoreFrom(inWindow) })
	if g.Turn.Step != StepFirstStrikeDamage {
		t.Fatalf("after the undo the cursor is at %q, want %q", g.Turn.Step, StepFirstStrikeDamage)
	}
	if got := StartingLife - g.Seats[1].Life; got != 2 {
		t.Fatalf("after the undo the defender has lost %d, want 2", got)
	}
	passUntilStep(t, g, StepCombatDamage)
	if got := StartingLife - g.Seats[1].Life; got != 4 {
		t.Errorf("after replaying the second step the defender lost %d, want 4 — not the first striker again", got)
	}
}

// A game saved and restored while paused between the steps comes back
// in the window with the record intact.
func TestSnapshotBetweenTheCombatDamageStepsRoundTrips(t *testing.T) {
	g := newActiveGame(t)
	atk := g.Seats[0]
	striker := pushDamageStepCreature(g, atk, "Youthful Knight", 2, 1, "first strike")
	bear := pushDamageStepCreature(g, atk, "Grizzly Bears", 2, 2)
	declareAttacks(t, g, striker, bear)

	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep into the first damage step: %v", err)
	}

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

	if restored.Turn.Step != StepFirstStrikeDamage {
		t.Fatalf("restored cursor at %q, want %q", restored.Turn.Step, StepFirstStrikeDamage)
	}
	if !restored.firstStrikeStepParticipants[striker] {
		t.Error("the restored game lost the first-strike step's participation record")
	}
	if restored.firstStrikeStepParticipants[bear] {
		t.Error("the Bears are in the restored participation record, but they never had first strike")
	}
	if got := StartingLife - restored.Seats[1].Life; got != 2 {
		t.Fatalf("restored defender has lost %d, want 2", got)
	}
	passUntilStep(t, restored, StepCombatDamage)
	if got := StartingLife - restored.Seats[1].Life; got != 4 {
		t.Errorf("after the restored second step the defender lost %d, want 4", got)
	}
}

// The record is one combat's bookkeeping: the end of combat step
// clears it with the declarations it describes as that step ends
// (CR 511.3), so the next combat starts from nothing.
func TestClearCombatForgetsTheParticipationRecord(t *testing.T) {
	g := newActiveGame(t)
	atk := g.Seats[0]
	striker := pushDamageStepCreature(g, atk, "Youthful Knight", 2, 1, "first strike")
	declareAttacks(t, g, striker)

	passUntilStep(t, g, StepCombatDamage)
	if len(g.firstStrikeStepParticipants) == 0 {
		t.Fatal("no participation record after a combat with first strike in it")
	}
	passUntilStep(t, g, StepEndCombat)
	if len(g.firstStrikeStepParticipants) == 0 {
		t.Errorf("the record was dropped on ENTRY to end_combat; combat lasts through the step (CR 511.3)")
	}
	passUntilStep(t, g, StepPostcombatMain)
	if len(g.firstStrikeStepParticipants) != 0 {
		t.Errorf("the record survived the end of combat step: %v", g.firstStrikeStepParticipants)
	}
}
