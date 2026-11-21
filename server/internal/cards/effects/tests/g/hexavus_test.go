package testsg

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const hexavusOracle = "ba48a1a6-a53f-4e7e-881a-472637a3a007"

func TestHexavusIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(hexavusOracle)
	if !ok || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 || !effectstest.HasKeyword(spec.PrintedKeywords, "flying") {
		t.Fatalf("registered %v, spec %+v", ok, spec)
	}
}

// Cast, it enters with six +1/+1 counters: a 6/6 flier, never a 0/0.
func TestHexavusEntersWithSixCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.CastHoldCreature(t, g, "Hexavus", "Artifact Creature — Construct", hexavusOracle, 0, 0)
	if got := effectstest.CountersOn(g, id, game.CounterPlusOne); got != 6 {
		t.Fatalf("%d +1/+1 counters, want 6", got)
	}
	if p, tt := effectstest.PtOf(t, g, id); p != 6 || tt != 6 {
		t.Errorf("%d/%d, want 6/6", p, tt)
	}
}

// {1}, remove a +1/+1 counter from Hexavus: a flying counter on another
// target creature.
func TestHexavusGivesAFlyingCounterByRemovingItsOwn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	hex := effectstest.B12Push(g, me.ID, "Hexavus", "Artifact Creature — Construct", hexavusOracle, 0, 0)
	effectstest.B28AddCounter(t, g, hex, 6)
	bear := effectstest.B12Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)

	// It cannot target itself.
	effectstest.ApaMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, hex, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: hex}}}); err == nil {
		t.Fatal("Hexavus targeted itself")
	}
	effectstest.ApaActivate(t, g, me, hex, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}})
	if got := effectstest.CountersOn(g, hex, game.CounterPlusOne); got != 5 {
		t.Errorf("Hexavus has %d +1/+1 counters, want 5", got)
	}
	if got := effectstest.CountersOn(g, bear, game.CounterFlying); got != 1 {
		t.Errorf("the target has %d flying counters, want 1", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the target does not fly")
	}
}

// {1}, remove a counter from another creature you control: a +1/+1
// counter on Hexavus. Its own counters cannot pay.
func TestHexavusMovesACounterFromAnotherCreatureOntoItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	hex := effectstest.B12Push(g, me.ID, "Hexavus", "Artifact Creature — Construct", hexavusOracle, 0, 0)
	effectstest.B28AddCounter(t, g, hex, 1)
	other := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	effectstest.B28AddCounter(t, g, other, 2)

	effectstest.ApaMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, hex, 1, game.ActivateAbilityParams{Strict: true,
		CounterSourceIDs: []uuid.UUID{hex}, CounterCounts: []int{1}, CounterKinds: []string{game.CounterPlusOne}}); err == nil {
		t.Fatal("paid with Hexavus's own counter")
	}
	effectstest.ApaActivate(t, g, me, hex, 1, game.ActivateAbilityParams{
		CounterSourceIDs: []uuid.UUID{other}, CounterCounts: []int{1}, CounterKinds: []string{game.CounterPlusOne}})
	if got := effectstest.CountersOn(g, other, game.CounterPlusOne); got != 1 {
		t.Errorf("the other creature has %d counters, want 1", got)
	}
	if got := effectstest.CountersOn(g, hex, game.CounterPlusOne); got != 2 {
		t.Errorf("Hexavus has %d counters, want 2", got)
	}
}
