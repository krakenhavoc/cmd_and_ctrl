package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const hexavusOracle = "ba48a1a6-a53f-4e7e-881a-472637a3a007"

func TestHexavusIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(hexavusOracle)
	if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 || !hasKeyword(spec.PrintedKeywords, "flying") {
		t.Fatalf("registered %v, spec %+v", ok, spec)
	}
}

// Cast, it enters with six +1/+1 counters: a 6/6 flier, never a 0/0.
func TestHexavusEntersWithSixCounters(t *testing.T) {
	g := newCatalogGame(t)
	id := castHoldCreature(t, g, "Hexavus", "Artifact Creature — Construct", hexavusOracle, 0, 0)
	if got := countersOn(g, id, game.CounterPlusOne); got != 6 {
		t.Fatalf("%d +1/+1 counters, want 6", got)
	}
	if p, tt := ptOf(t, g, id); p != 6 || tt != 6 {
		t.Errorf("%d/%d, want 6/6", p, tt)
	}
}

// {1}, remove a +1/+1 counter from Hexavus: a flying counter on another
// target creature.
func TestHexavusGivesAFlyingCounterByRemovingItsOwn(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	hex := b12Push(g, me.ID, "Hexavus", "Artifact Creature — Construct", hexavusOracle, 0, 0)
	b28AddCounter(t, g, hex, 6)
	bear := b12Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)

	// It cannot target itself.
	apaMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, hex, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: hex}}}); err == nil {
		t.Fatal("Hexavus targeted itself")
	}
	apaActivate(t, g, me, hex, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}})
	if got := countersOn(g, hex, game.CounterPlusOne); got != 5 {
		t.Errorf("Hexavus has %d +1/+1 counters, want 5", got)
	}
	if got := countersOn(g, bear, game.CounterFlying); got != 1 {
		t.Errorf("the target has %d flying counters, want 1", got)
	}
	if !hasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the target does not fly")
	}
}

// {1}, remove a counter from another creature you control: a +1/+1
// counter on Hexavus. Its own counters cannot pay.
func TestHexavusMovesACounterFromAnotherCreatureOntoItself(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	hex := b12Push(g, me.ID, "Hexavus", "Artifact Creature — Construct", hexavusOracle, 0, 0)
	b28AddCounter(t, g, hex, 1)
	other := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	b28AddCounter(t, g, other, 2)

	apaMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, hex, 1, game.ActivateAbilityParams{Strict: true,
		CounterSourceIDs: []uuid.UUID{hex}, CounterCounts: []int{1}, CounterKinds: []string{game.CounterPlusOne}}); err == nil {
		t.Fatal("paid with Hexavus's own counter")
	}
	apaActivate(t, g, me, hex, 1, game.ActivateAbilityParams{
		CounterSourceIDs: []uuid.UUID{other}, CounterCounts: []int{1}, CounterKinds: []string{game.CounterPlusOne}})
	if got := countersOn(g, other, game.CounterPlusOne); got != 1 {
		t.Errorf("the other creature has %d counters, want 1", got)
	}
	if got := countersOn(g, hex, game.CounterPlusOne); got != 2 {
		t.Errorf("Hexavus has %d counters, want 2", got)
	}
}
