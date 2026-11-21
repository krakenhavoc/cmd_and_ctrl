package testss82v

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// A Fog is a restore point, and the restored game still prevents
// combat damage.
func TestFogIsARestorePointAndStillPrevents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0].ID
	effectstest.CastCatalogSpell(t, g, "Fog", "Instant", effectstest.FogOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	restored := effectstest.RestoreRoundTrip(t, g, true)
	if n := effectstest.ScopedReplacementCount(restored); n != 1 {
		t.Fatalf("restored game holds %d scoped replacements, want the Fog", n)
	}
	defender := effectstest.PushBear(restored, me, "Defender", 2)
	attacker := effectstest.PushBear(restored, restored.Seats[1].ID, "Attacker", 2)
	if err := restored.MarkCombatDamage(attacker, defender, 2); err != nil {
		t.Fatalf("MarkCombatDamage: %v", err)
	}
	if got := effectstest.DamageMarkedOn(restored, defender); got != 0 {
		t.Errorf("combat damage under a restored Fog: marked %d, want 0", got)
	}
}

// A half-spent Mending Hands keeps exactly the charge it had left
// across a restore.
func TestMendingHandsPartialChargeSurvivesARestore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0].ID
	bear := effectstest.PushBear(g, me, "Shielded Bear", 10)
	effectstest.CastCatalogSpell(t, g, "Mending Hands", "Instant", effectstest.MendingHandsOracl,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.New(), bear, 3) })
	if got := effectstest.DamageMarkedOn(g, bear); got != 0 {
		t.Fatalf("setup: 3 into a 4-shield marked %d", got)
	}

	restored := effectstest.RestoreRoundTrip(t, g, true)
	restored.WithWriteLock(func() { _ = restored.DealDamageToCreatureForEffect(uuid.New(), bear, 3) })
	if got := effectstest.DamageMarkedOn(restored, bear); got != 2 {
		t.Errorf("restored: 3 into the 1 charge left marked %d, want 2", got)
	}
	if n := effectstest.ScopedReplacementCount(restored); n != 0 {
		t.Errorf("a spent shield is removed; %d scoped replacements left", n)
	}
}

// The behaviour change ADR 0041 P8 calls out: an undo rewinds a
// shield's spent charge. The charge used to be a variable the closure
// captured, which the undo did not reach.
func TestUndoRewindsASpentShield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0].ID
	bear := effectstest.PushBear(g, me, "Shielded Bear", 10)
	effectstest.CastCatalogSpell(t, g, "Mending Hands", "Instant", effectstest.MendingHandsOracl,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)

	snap := g.Clone()
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.New(), bear, 3) })
	g.WithWriteLock(func() { g.RestoreFrom(snap) })

	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.New(), bear, 4) })
	if got := effectstest.DamageMarkedOn(g, bear); got != 0 {
		t.Errorf("after the undo, 4 into the shield marked %d, want 0 — the undo rewinds the charge", got)
	}
	// And the snapshot the undo came from still holds the full charge:
	// nothing wrote through a shared record.
	for _, e := range snap.ScopedEffects {
		for _, m := range e.Mods {
			if m.Kind == game.ModPreventDamage && m.Amount != 4 {
				t.Errorf("the undo snapshot's shield reads %d, want 4 — a record was written in place", m.Amount)
			}
		}
	}
}

// A shield on a permanent is pinned to that object (CR 400.7): once
// the permanent is gone the record goes with it.
func TestMendingHandsShieldEndsWithItsPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0].ID
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Shielded Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 10, Owner: me, Controller: me,
	})
	effectstest.CastCatalogSpell(t, g, "Mending Hands", "Instant", effectstest.MendingHandsOracl,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.ScopedReplacementCount(g); n != 1 {
		t.Fatalf("setup: %d scoped replacements, want 1", n)
	}
	g.WithWriteLock(func() {
		_ = g.SacrificePermanentForEffect(bear)
		g.RecomputeLayersIfStaleLocked()
	})
	if n := effectstest.ScopedReplacementCount(g); n != 0 {
		t.Errorf("the shield outlived its permanent: %d scoped replacements", n)
	}
}

// Druid's Deliverance prevents combat damage to its controller and to
// nothing else.
func TestDruidsDeliveranceShieldsOnlyItsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.CastCatalogSpell(t, g, "Druid's Deliverance", "Instant", "fde7645a-5f02-4d5f-b38c-8390f325899e", nil)
	effectstest.PassPriorityAroundTable(t, g)

	mine := effectstest.PushBear(g, me.ID, "My Blocker", 5)
	theirs := effectstest.PushBear(g, opp.ID, "Their Attacker", 5)
	if err := g.MarkCombatDamage(theirs, mine, 2); err != nil {
		t.Fatalf("MarkCombatDamage: %v", err)
	}
	if got := effectstest.DamageMarkedOn(g, mine); got != 2 {
		t.Errorf("combat damage to the controller's creature: marked %d, want 2 — only damage to the player is prevented", got)
	}
}

// #1591. The redirect lasts as long as the returned object is on the
// battlefield — past cleanup when the end-step exile was countered —
// so a creature that survived its turn still goes to exile, not the
// graveyard, and cannot be whipped twice. It is data, so the table is a
// restore point while it holds it.
func TestWhipRedirectOutlivesACounteredExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	effectstest.AdvanceToMainOf(t, g, seat)
	whip := effectstest.PushCatalogPermanent(g, me.ID, "Whip of Erebos", "Legendary Enchantment Artifact", effectstest.B06WhipOfErebosOracle, false)
	dead := effectstest.SeedGraveyardCreature(me, "Giant", "{4}{B}")
	effectstest.B06AddMana(me, "B", "B", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, whip, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: dead}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dead) {
		t.Fatal("setup: the creature should be back")
	}

	// Stifle the end-step exile.
	effectstest.AdvanceToEndStepOf(t, g, seat)
	item := effectstest.TriggerOnStack(g, whip)
	if item == nil {
		t.Fatal("setup: the delayed exile is not on the stack at the end step")
	}
	if err := g.CounterAbility(item.ID); err != nil {
		t.Fatalf("CounterAbility: %v", err)
	}

	// Into the next turn: past the cleanup that used to end the redirect.
	effectstest.AdvanceToMainOf(t, g, (seat+1)%len(g.Seats))
	if !g.Battlefield.Contains(dead) {
		t.Fatal("setup: the creature should have survived its turn")
	}
	if n := effectstest.ScopedReplacementCount(g); n != 1 {
		t.Fatalf("the redirect did not survive cleanup: %d scoped replacements", n)
	}

	restored := effectstest.RestoreRoundTrip(t, g, true)
	restored.WithWriteLock(func() { _ = restored.SacrificePermanentForEffect(dead) })
	owner := restored.Seats[seat]
	if owner.Graveyard.Contains(dead) {
		t.Error("the whipped creature reached the graveyard a turn later (#1591)")
	}
	if !restored.Exile.Contains(dead) {
		t.Error("it is exiled instead")
	}
	restored.WithWriteLock(func() { restored.RecomputeLayersIfStaleLocked() })
	if n := effectstest.ScopedReplacementCount(restored); n != 0 {
		t.Errorf("the redirect outlived its object: %d scoped replacements", n)
	}
}

// Unearth shares the redirect, and a bounce is an exit like any other
// (#539): the unearthed creature goes to exile, not to hand.
func TestUnearthedCreatureBouncedGoesToExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id, me := effectstest.ActivateFromGraveyard(t, g, "Dregscape Zombie", "Creature — Zombie",
		effectstest.DregscapeZombieOracle, 2, 1, "{B}", game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.BounceToHandForEffect(id); err != nil {
		t.Fatalf("BounceToHandForEffect: %v", err)
	}
	if me.Hand.Contains(id) {
		t.Error("the unearthed Zombie went to hand — it must be exiled")
	}
	if !g.Exile.Contains(id) {
		t.Error("the unearthed Zombie is not in exile")
	}
}

// Cosmic Intervention is a restore point, and the restored game still
// exiles instead and schedules the return.
func TestCosmicInterventionIsARestorePointAndStillExiles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Cosmic Intervention", "Instant", effectstest.CosmicInterventionOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	restored := effectstest.RestoreRoundTrip(t, g, true)
	// Read live: a permanent that arrived after the spell resolved is
	// saved too.
	drifter := effectstest.PushFlickerCreature(restored, me.ID, "Mulldrifter", effectstest.MulldrifterOracle)
	restored.WithWriteLock(func() {
		if err := restored.DestroyPermanentForEffect(drifter); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
	if !restored.Exile.Contains(drifter) {
		t.Fatal("the restored Cosmic Intervention did not exile the permanent")
	}
	if len(restored.DelayedTriggers) != 1 {
		t.Fatalf("delayed-return queue holds %d triggers, want 1", len(restored.DelayedTriggers))
	}
	if got := restored.DelayedTriggers[0].Label; got != "Cosmic Intervention — return the exiled permanent" {
		t.Errorf("delayed return labelled %q", got)
	}
}
