package game

import (
	"slices"
	"testing"

	"github.com/google/uuid"
)

// redirect_order_instance_test.go — #2066: the CR 616 ordering of two
// redirections is asked once per damage instance.

// openReplacementOrder is the open CR 616 ordering prompt, nil when none.
func openReplacementOrder(g *Game) *PendingChoice {
	var out *PendingChoice
	g.WithWriteLock(func() {
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == PendingChoiceReplacementOrder {
				out = c
				return
			}
		}
	})
	return out
}

// Gideon's Sacrifice / Saving Grace ruling: when two such redirections
// meet simultaneous damage, the player chooses one creature and ALL of
// that damage goes to it. The CR 616 ordering is therefore asked once for
// the damage instance, not once per event in it: two attackers' combat
// damage is one instance, and the answer given for the first applies to
// the second.
func TestRedirectOrderIsChosenOncePerDamageInstance(t *testing.T) {
	g := newActiveGame(t)
	atk, def := g.Seats[0], g.Seats[1]
	a := pushKeywordCreature(t, g, atk, 3, 3)
	b := pushKeywordCreature(t, g, atk, 2, 2)
	x := pushKeywordCreature(t, g, def, 1, 20)
	y := pushKeywordCreature(t, g, def, 1, 20)
	g.WithWriteLock(func() {
		for _, to := range []uuid.UUID{x, y} {
			if !g.RedirectDamageThisTurnForEffect(DamageRedirection{Controller: def.ID, ProtectPlayer: def.ID, To: to, Label: "Saving Grace"}) {
				t.Fatal("no redirection")
			}
		}
	})
	defStart := lifeOf(g, def.ID)
	advanceIntoStep(t, g, StepDeclareAttackers)
	for _, c := range []uuid.UUID{a, b} {
		if err := g.DeclareAttacker(c, def.ID); err != nil {
			t.Fatal(err)
		}
	}
	advanceIntoStep(t, g, StepCombatDamage)
	prompts := 0
	for c := openReplacementOrder(g); c != nil; c = openReplacementOrder(g) {
		prompts++
		if prompts > 1 {
			t.Fatalf("a second ordering prompt for the same damage instance")
		}
		ids := slices.Clone(c.ReplacementEffectIDs)
		slices.Reverse(ids)
		if err := g.ResolveReplacementOrder(c.ID, c.Chooser, ids); err != nil {
			t.Fatal(err)
		}
	}
	if prompts != 1 {
		t.Fatalf("%d ordering prompts, want exactly one", prompts)
	}
	if lifeOf(g, def.ID) != defStart {
		t.Fatalf("defender lost life: every event was redirected")
	}
	mx, my := markedOn(g, x), markedOn(g, y)
	if !(mx == 5 && my == 0) && !(mx == 0 && my == 5) {
		t.Fatalf("damage split x=%d y=%d: all of it must go to the one creature chosen", mx, my)
	}
}

// An event of the instance that arrives AFTER the answer (a damage walk
// that carried on once the first event's prompt was answered) takes the
// remembered order instead of asking again.
func TestRedirectOrderAnsweredEarlierStandsForALaterEvent(t *testing.T) {
	g := newActiveGame(t)
	atk, def := g.Seats[0], g.Seats[1]
	src := pushKeywordCreature(t, g, atk, 3, 3)
	x := pushKeywordCreature(t, g, def, 1, 20)
	y := pushKeywordCreature(t, g, def, 1, 20)
	g.WithWriteLock(func() {
		for _, to := range []uuid.UUID{x, y} {
			if !g.RedirectDamageThisTurnForEffect(DamageRedirection{Controller: def.ID, ProtectPlayer: def.ID, To: to, Label: "Saving Grace"}) {
				t.Fatal("no redirection")
			}
		}
	})
	var inst DamageInstance
	g.WithWriteLock(func() {
		inst = g.nextDamageInstanceLocked()
		g.openDamageInstance = inst
		_ = g.DealDamageToPlayerForEffect(src, def.ID, 2)
		g.openDamageInstance = 0
	})
	c := openReplacementOrder(g)
	if c == nil {
		t.Fatal("no ordering prompt for two redirections")
	}
	ids := slices.Clone(c.ReplacementEffectIDs)
	slices.Reverse(ids)
	if err := g.ResolveReplacementOrder(c.ID, c.Chooser, ids); err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() {
		g.openDamageInstance = inst
		_ = g.DealDamageToPlayerForEffect(src, def.ID, 3)
		g.openDamageInstance = 0
	})
	if openReplacementOrder(g) != nil {
		t.Fatal("asked the order again for a later event of the same instance")
	}
	mx, my := markedOn(g, x), markedOn(g, y)
	if !(mx == 5 && my == 0) && !(mx == 0 && my == 5) {
		t.Fatalf("damage split x=%d y=%d: all of it must go to the one creature chosen", mx, my)
	}
}
