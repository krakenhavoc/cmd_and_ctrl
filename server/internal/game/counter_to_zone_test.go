package game

import (
	"testing"

	"github.com/google/uuid"
)

// counter_to_zone_test.go — #1230: counterSpellLocked already accepted
// a destination zone (proven at the CounterSpell admin-API layer by
// TestS131CounterSpellHonoursDestination in mutations_test.go); the
// missing piece was a card-effect entry point for it
// (CounterTargetToZoneForEffect) and a SIBLING primitive for the
// "return target spell to its owner's hand" shape that never counts
// as a counter at all (ReturnSpellToHandForEffect — Reprieve,
// Narset's Reversal). Both share counterSpellLocked's body
// (exitSpellFromStackLocked), so this file pins what differs between
// them rather than re-proving what commander_zone_routes_test.go and
// mutations_test.go already cover.

// TestCounterTargetToZoneForEffectHonoursDestination is the card-
// effect wrapper's own proof, mirroring TestS131CounterSpellHonoursDestination
// one level up the stack.
func TestCounterTargetToZoneForEffectHonoursDestination(t *testing.T) {
	g := newActiveGame(t)
	caster := g.Seats[0]
	id := seatCommander(t, g.Stack, caster) // any spell; commander-ness unused here
	top, _ := g.Stack.Top()
	_, _ = g.Stack.Remove(id)
	pushStackSpell(t, g, top)

	g.mu.Lock()
	err := g.CounterTargetToZoneForEffect(id, ZoneRef{Kind: ZoneExile})
	g.mu.Unlock()
	// The commander is still asked CR 903.9 before landing in exile —
	// answer it so the assertion below can read a settled zone.
	if err != nil {
		t.Fatalf("CounterTargetToZoneForEffect: %v", err)
	}
	prompt := expectCommanderPrompt(t, g, caster)
	if err := g.ResolveOptionalReplacement(prompt.ID, caster.ID, false); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	if !g.Exile.Contains(id) {
		t.Error("CounterTargetToZoneForEffect should have exiled the spell")
	}
	if !hasEventFor(g, EventCounterSpell, id) {
		t.Error("this IS a counter (CR 701.6a) — EventCounterSpell should fire")
	}
}

// TestReturnSpellToHandForEffectIsNotACounter is the whole reason the
// primitive exists beside counterSpellLocked: no EventCounterSpell,
// whatever the catalog says about "can't be countered" — the check
// lives in CounterTargetForEffect / CounterTargetToZoneForEffect, and
// ReturnSpellToHandForEffect never consults it.
func TestReturnSpellToHandForEffectIsNotACounter(t *testing.T) {
	g := newActiveGame(t)
	caster := g.Seats[0]
	id := uuid.MustParse("00000000-0000-0000-0000-0000000000e2")
	pushStackSpell(t, g, Card{
		InstanceID: id, Name: "Uncounterable Spell", OracleID: "test-cant-be-countered",
		Owner: caster.ID, Controller: caster.ID,
	})
	old := CatalogCantBeCountered
	CatalogCantBeCountered = func(oracleID string) bool { return oracleID == "test-cant-be-countered" }
	t.Cleanup(func() { CatalogCantBeCountered = old })

	g.mu.Lock()
	err := g.ReturnSpellToHandForEffect(id)
	g.mu.Unlock()
	if err != nil {
		t.Fatalf("ReturnSpellToHandForEffect: %v", err)
	}
	if g.Stack.Contains(id) {
		t.Error("the spell should be off the stack")
	}
	if !caster.Hand.Contains(id) {
		t.Error("a spell that can't be countered is still returned to hand — CR 701.6 does not apply")
	}
	if hasEventFor(g, EventCounterSpell, id) {
		t.Error("EventCounterSpell must not fire — this is not a counter")
	}
	if _, ok := g.StackMeta[id]; ok {
		t.Error("StackMeta entry should be dropped, same as a counter")
	}
}

// TestCommanderReturnedToHandOffersCommandZone: CR 903.9's "from
// anywhere" reaches ReturnSpellToHandForEffect's stack exit exactly
// as it reaches counterSpellLocked's.
func TestCommanderReturnedToHandOffersCommandZone(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	cmdID := seatCommander(t, g.Stack, owner)
	top, _ := g.Stack.Top()
	_, _ = g.Stack.Remove(cmdID)
	pushStackSpell(t, g, top)

	g.mu.Lock()
	err := g.ReturnSpellToHandForEffect(cmdID)
	g.mu.Unlock()
	if err != nil {
		t.Fatalf("ReturnSpellToHandForEffect: %v", err)
	}
	if owner.Hand.Contains(cmdID) {
		t.Fatalf("returned commander reached hand before the prompt was answered")
	}
	prompt := expectCommanderPrompt(t, g, owner)
	if err := g.ResolveOptionalReplacement(prompt.ID, owner.ID, true); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	assertOnlyIn(t, cmdID, owner.Command, owner.Hand, g.Stack)
}

// TestCommanderReturnedToHandDeclineGoesToHand: declining still
// returns the spell to hand, same as a countered one would to the
// graveyard.
func TestCommanderReturnedToHandDeclineGoesToHand(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	cmdID := seatCommander(t, g.Stack, owner)
	top, _ := g.Stack.Top()
	_, _ = g.Stack.Remove(cmdID)
	pushStackSpell(t, g, top)

	g.mu.Lock()
	err := g.ReturnSpellToHandForEffect(cmdID)
	g.mu.Unlock()
	if err != nil {
		t.Fatalf("ReturnSpellToHandForEffect: %v", err)
	}
	prompt := expectCommanderPrompt(t, g, owner)
	if err := g.ResolveOptionalReplacement(prompt.ID, owner.ID, false); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	assertOnlyIn(t, cmdID, owner.Hand, owner.Command, g.Stack)
}
