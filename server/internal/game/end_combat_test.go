package game

import (
	"testing"

	"github.com/google/uuid"
)

// end_combat_test.go — CR 724.2, "end the combat phase" (#1317).
// The resolver is stubbed (effect_hooks_test.go's withEffectHooks), so
// these tests drive the engine verb through a real resolution without
// importing the catalog. The catalog's own card test is
// effects/mandate_of_peace_test.go.

const endCombatStubOracle = "stub-end-the-combat-phase"

// withEndCombatResolver installs a resolver under which a spell with
// endCombatStubOracle ends the combat phase as it resolves.
func withEndCombatResolver(t *testing.T) {
	t.Helper()
	withEffectHooks(t,
		func(g *Game, item *StackItem, oracleID string) error {
			if oracleID == endCombatStubOracle {
				g.EndCombatPhaseForEffect()
			}
			return nil
		},
		nil,
		func(oracleID string) bool { return oracleID == endCombatStubOracle },
	)
}

// castEndCombatSpell puts an instant carrying the stub oracle in
// caster's hand and casts it.
func castEndCombatSpell(t *testing.T, g *Game, caster *Player) uuid.UUID {
	t.Helper()
	id := pushTypedCardToHand(caster, "Mandate Stub", "Instant")
	for i := range caster.Hand.Cards {
		if caster.Hand.Cards[i].InstanceID == id {
			caster.Hand.Cards[i].OracleID = endCombatStubOracle
		}
	}
	if err := g.CastSpell(caster.ID, id, CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	return id
}

// passUntilStackEmpty passes priority until nothing is on the stack.
func passUntilStackEmpty(t *testing.T, g *Game) {
	t.Helper()
	for range 32 {
		if len(g.Stack.Cards) == 0 && len(g.StackMeta) == 0 {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatalf("the stack never emptied")
}

func stackItems(g *Game) []*StackItem {
	out := make([]*StackItem, 0, len(g.StackMeta))
	for _, item := range g.StackMeta {
		out = append(out, item)
	}
	return out
}

func inExile(g *Game, id uuid.UUID) bool {
	return g.Exile != nil && g.Exile.Contains(id)
}

// The whole of CR 724.2 in one resolution: every object on the stack
// is exiled — the resolving spell, a spell beneath it and an ability
// beneath that — every creature leaves combat, the end of combat step
// never begins, and the active player receives priority in the
// postcombat main phase.
func TestEndCombatPhaseExilesTheStackAndSkipsToPostcombatMain(t *testing.T) {
	withEndCombatResolver(t)
	g := newActiveGame(t)
	attacker := pushCombatant(t, g, g.Seats[0], "Attacker", 3, 3)
	blocker := pushCombatant(t, g, g.Seats[1], "Blocker", 1, 1)
	blockAfterLockIn(t, g, attacker, blocker)

	// Beneath: an ordinary instant, then an activated ability.
	active := g.Seats[0]
	under := pushTypedCardToHand(active, "Lesser Instant", "Instant")
	if err := g.CastSpell(active.ID, under, CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	abilityRan := false
	abilityID := uuid.New()
	g.WithWriteLock(func() {
		g.StackMeta[abilityID] = &StackItem{
			ID:           abilityID,
			Kind:         StackItemActivated,
			Controller:   active.ID,
			Owner:        active.ID,
			SourceCardID: attacker,
			Label:        "an ability beneath",
			Seq:          g.nextStackSeqLocked(),
			Effect: func(*Game, *StackItem) error {
				abilityRan = true
				return nil
			},
		}
	})
	mandate := castEndCombatSpell(t, g, active)

	seq := lastSeq(g)
	passUntilStackEmpty(t, g)

	if g.Turn.Step != StepPostcombatMain {
		t.Fatalf("the cursor is at %s, want postcombat_main (CR 724.2d)", g.Turn.Step)
	}
	if g.Turn.PriorityHolder != g.Turn.ActiveSeat {
		t.Errorf("priority is with seat %d, want the active seat %d", g.Turn.PriorityHolder, g.Turn.ActiveSeat)
	}
	if !inExile(g, mandate) {
		t.Errorf("the resolving spell was not exiled (CR 724.2b: \"including the object that's resolving\")")
	}
	if !inExile(g, under) {
		t.Errorf("the spell beneath it was not exiled (CR 724.2b)")
	}
	if len(g.StackMeta) != 0 || len(g.Stack.Cards) != 0 {
		t.Errorf("the stack kept %d records and %d cards; exiling must clear StackMeta", len(g.StackMeta), len(g.Stack.Cards))
	}
	if abilityRan {
		t.Errorf("the ability beneath resolved; it should have been exiled with the stack")
	}
	for _, c := range g.Battlefield.Cards {
		if c.AttackingTarget != uuid.Nil || c.BlockingTarget != uuid.Nil {
			t.Errorf("%s is still in combat (CR 724.2d)", c.Name)
		}
	}
	if stepBeganSeq(t, g, StepEndCombat) != 0 || stepBeganSeq(t, g, StepCombatDamage) != 0 {
		t.Errorf("a skipped combat step began (CR 724.2d / 724.2e)")
	}
	for _, ev := range g.Events {
		if ev.Seq <= seq {
			continue
		}
		if ev.Kind == EventCounterSpell {
			t.Errorf("an exiled stack object was reported countered: %+v", ev)
		}
		if ev.Kind == EventDealDamage && ev.Combat {
			t.Errorf("combat damage was dealt after the combat phase ended: %+v", ev)
		}
	}
	if findCard(g, blocker) == nil {
		t.Errorf("the blocker died; no combat damage should have been dealt")
	}
}

// CR 724.2g: outside a combat phase the instruction does nothing, and
// the spell resolves into the graveyard like any other instant.
func TestEndCombatPhaseOutsideCombatDoesNothing(t *testing.T) {
	withEndCombatResolver(t)
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	active := g.Seats[0]
	id := castEndCombatSpell(t, g, active)
	passUntilStackEmpty(t, g)

	if g.Turn.Step != StepPrecombatMain {
		t.Errorf("the cursor moved to %s outside combat (CR 724.2g)", g.Turn.Step)
	}
	if inExile(g, id) || !active.Graveyard.Contains(id) {
		t.Errorf("the spell did not resolve into its owner's graveyard")
	}
}

// CR 724.2a: triggers that triggered but were not yet put on the stack
// cease to exist. CR 724.2e: an "at end of combat" delayed trigger does
// not fire, because the end of combat step never begins — it stays
// queued for the next one.
func TestEndCombatPhaseDropsPendingTriggersAndSkipsEndOfCombat(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushCombatant(t, g, g.Seats[0], "Attacker", 2, 2)
	declareAttacks(t, g, attacker)

	fired := false
	g.WithWriteLock(func() {
		g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
			Controller: g.Seats[0].ID,
			Label:      "at end of combat",
			At:         StepEndCombat,
			Effect: func(*Game, *StackItem) error {
				fired = true
				return nil
			},
		})
		g.PendingTriggers = append(g.PendingTriggers, &StackItem{
			ID: uuid.New(), Kind: StackItemTriggered, Controller: g.Seats[0].ID,
			Label: "pending before the process",
		})
		g.EndCombatPhaseForEffect()
	})
	passUntilStackEmpty(t, g)

	if g.Turn.Step != StepPostcombatMain {
		t.Fatalf("at %s, want postcombat_main", g.Turn.Step)
	}
	for _, item := range append(append([]*StackItem(nil), g.PendingTriggers...), stackItems(g)...) {
		if item != nil && item.Label == "pending before the process" {
			t.Errorf("a pending trigger survived the process (CR 724.2a)")
		}
	}
	if fired {
		t.Errorf("an \"at end of combat\" delayed trigger fired (CR 724.2e)")
	}
	queued := false
	for _, dt := range g.DelayedTriggers {
		if dt != nil && dt.At == StepEndCombat {
			queued = true
		}
	}
	if !queued {
		t.Errorf("the end-of-combat delayed trigger was dropped; it waits for the next end of combat step")
	}
}

// A copy of a spell on the stack is not a card (CR 707.10): it goes
// nowhere and its record goes with it.
func TestEndCombatPhaseCopyOnTheStackCeasesToExist(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushCombatant(t, g, g.Seats[0], "Attacker", 2, 2)
	declareAttacks(t, g, attacker)
	active := g.Seats[0]
	orig := pushTypedCardToHand(active, "Original", "Instant")
	if err := g.CastSpell(active.ID, orig, CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	g.WithWriteLock(func() {
		g.StackMeta[orig].IsCopy = true
		g.EndCombatPhaseForEffect()
	})
	if inExile(g, orig) || g.Stack.Contains(orig) || g.StackMeta[orig] != nil {
		t.Errorf("a copy was exiled or left on the stack instead of ceasing to exist")
	}
}
