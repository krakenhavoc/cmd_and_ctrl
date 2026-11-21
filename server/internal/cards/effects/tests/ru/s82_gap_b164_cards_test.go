package testsru

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b164_cards_test.go — Jumpstart 2022 creatures, GAP batch 164 (#5948).
// Every test plays the real catalog entry through the engine.

const (
	gap164HonorGuard = "0595b074-8f25-469d-9e82-ed60c4a7a2cd"
	gap164Professor  = "1b59665f-23c7-4df9-8e71-f9b1a419063b"
	gap164Aerosaur   = "6889a555-6fde-4728-9cf2-715f0730efb7"
)

func TestGapB164EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{gap164HonorGuard, gap164Professor, gap164Aerosaur} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

func TestGapB164HonorGuardGetsPlusOneForEachLegendaryCreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	guard := effectstest.B12Push(g, me.ID, "Benalish Honor Guard", "Creature — Human Knight", gap164HonorGuard, 2, 2)
	if got := effectstest.EffectivePower(t, g, guard); got != 2 {
		t.Fatalf("power %d with no legend, want 2", got)
	}
	legend := effectstest.B12Creature(g, me.ID, "Legend", "Legendary Creature — Human", 1, 1)
	effectstest.B12Creature(g, opp.ID, "Their Legend", "Legendary Creature — Human", 1, 1)
	effectstest.B12Permanent(g, me.ID, "Legendary Rock", "Legendary Artifact")
	if got := effectstest.EffectivePower(t, g, guard); got != 3 {
		t.Errorf("power %d with one legendary creature (opponent's creature and a noncreature legend don't count), want 3", got)
	}
	effectstest.B12Creature(g, me.ID, "Second Legend", "Legendary Creature — Elf", 1, 1)
	if got := effectstest.EffectivePower(t, g, guard); got != 4 {
		t.Errorf("power %d with two legendary creatures, want 4", got)
	}
	effectstest.B27Kill(g, legend)
	g.RunStateChecksForTest()
	if got := effectstest.EffectivePower(t, g, guard); got != 3 {
		t.Errorf("power %d after a legend died, want 3", got)
	}
	if got := effectstest.ApaLive(g, guard).CurrentToughness(); got != 2 {
		t.Errorf("toughness %d, want it unchanged at 2", got)
	}
}

func TestGapB164CombatProfessorPumpsATargetAtBeginningOfCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	prof := effectstest.B12Push(g, me.ID, "Combat Professor", "Creature — Bird Cleric", gap164Professor, 2, 3)
	if !effectstest.HasKeywordOnBattlefield(t, g, prof, "flying") {
		t.Error("the Professor lacks flying")
	}
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if effectstest.LatestPickTarget(g, me.ID) == nil {
		t.Fatal("no target prompt at the beginning of combat")
	}
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, bear); got != 3 {
		t.Errorf("power %d, want 3", got)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "vigilance") {
		t.Error("the Bear did not gain vigilance")
	}
}

func TestGapB164CombatProfessorDoesNothingOnAnOpponentsTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Combat Professor", "Creature — Bird Cleric", gap164Professor, 2, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { g.Turn.ActiveSeat = 1; g.Turn.PriorityHolder = 1 })
	effectstest.AdvanceToMain(t, g)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("the Professor triggered on an opponent's turn")
	}
	if got := effectstest.EffectivePower(t, g, bear); got != 2 {
		t.Errorf("power %d, want 2", got)
	}
}

func TestGapB164ImperialAerosaurBoostsAnotherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	saur := effectstest.CastHoldCreature(t, g, "Imperial Aerosaur", "Creature — Dinosaur", gap164Aerosaur, 3, 3)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	for _, bad := range []game.TargetRef{{Kind: game.TargetCard, ID: saur}, {Kind: game.TargetCard, ID: theirs}} {
		if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{bad}); err == nil {
			t.Fatalf("target %v was accepted", bad.ID)
		}
	}
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	live := effectstest.ApaLive(g, bear)
	if live.CurrentPower() != 3 || live.CurrentToughness() != 3 {
		t.Errorf("P/T %d/%d, want 3/3", live.CurrentPower(), live.CurrentToughness())
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "flying") {
		t.Error("the Bear did not gain flying")
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, saur, "flying") {
		t.Error("the Aerosaur lacks flying")
	}
}

func TestGapB164ImperialAerosaurWithNoOtherCreatureDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastHoldCreature(t, g, "Imperial Aerosaur", "Creature — Dinosaur", gap164Aerosaur, 3, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("a target prompt was offered with no other creature")
	}
}
