package testss82v

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sin_unending_cataclysm_test.go — #3246 / #1893, ADR 0170.

const sinUnendingCataclysmOracle = "48534bbf-06b6-46a0-8a86-b14f38caee58"

func castSin(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	id := effectstest.CastCatalogSpell(t, g, "Sin, Unending Cataclysm", "Legendary Creature — Leviathan Avatar", sinUnendingCataclysmOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func TestSinStripsTheChosenAndEntersWithTwiceAsMany(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Hydra", "Creature — Hydra", 0, 0)
	effectstest.AddCounters(t, g, theirs, game.CounterPlusOne, 3)
	withered := effectstest.B12Creature(g, me.ID, "Withered Bear", "Creature — Bear", 3, 3)
	effectstest.AddCounters(t, g, withered, game.CounterMinusOne, 1)
	grown := effectstest.B12Creature(g, me.ID, "Grown Bear", "Creature — Bear", 2, 2)
	effectstest.AddCounters(t, g, grown, game.CounterPlusOne, 2)
	plain := effectstest.B12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	land := effectstest.B12Permanent(g, opp.ID, "Charged Land", "Land")
	effectstest.AddCounters(t, g, land, "charge", 4)
	saga := effectstest.B12Permanent(g, opp.ID, "Their Saga", "Enchantment — Saga")
	effectstest.AddCounters(t, g, saga, "lore", 1)

	id := castSin(t, g)
	c := effectstest.EntryChoiceOfKind(g, game.PendingChoiceEntryChoosePermanents, me.ID)
	if c == nil {
		t.Fatal("Sin did not ask")
	}
	if _, ok := effectstest.BattlefieldCard(g, id); ok {
		t.Fatal("Sin entered before the choice (CR 614.12)")
	}
	offered := map[uuid.UUID]bool{}
	for _, cand := range c.ChooseCards {
		offered[cand] = true
	}
	for _, want := range []uuid.UUID{theirs, withered, grown, saga} {
		if !offered[want] {
			t.Errorf("a countered artifact, creature or enchantment was not offered: %v", want)
		}
	}
	if offered[plain] || offered[land] {
		t.Error("offered a permanent with no counters, or a land")
	}
	if c.ChooseMin != 0 || c.ChooseMax != 4 {
		t.Errorf("bounds %d..%d, want 0..4 (any number)", c.ChooseMin, c.ChooseMax)
	}
	suggested := map[uuid.UUID]bool{}
	for _, s := range c.ChooseSuggested {
		suggested[s] = true
	}
	if !suggested[theirs] || !suggested[withered] || !suggested[saga] || suggested[grown] || len(suggested) != 3 {
		t.Errorf("suggested %v: want the opponent's two and my -1/-1 creature", c.ChooseSuggested)
	}

	effectstest.AnswerEntryChoice(t, g, game.PendingChoiceEntryChoosePermanents, me.ID, theirs, withered)
	for _, stripped := range []uuid.UUID{theirs, withered} {
		if n := len(effectstest.B12Card(t, g, stripped).Counters); n != 0 {
			t.Errorf("a chosen permanent kept %d counter kinds", n)
		}
	}
	if got := effectstest.B12Counter(t, g, grown, game.CounterPlusOne); got != 2 {
		t.Errorf("an unchosen creature lost its counters: %d", got)
	}
	if got := effectstest.B12Counter(t, g, id, game.CounterPlusOne); got != 8 {
		t.Errorf("Sin entered with %d +1/+1 counters, want 8 (twice the four removed)", got)
	}
}

func TestSinChoosingNothingEntersWithNoCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	grown := effectstest.B12Creature(g, me.ID, "Grown Bear", "Creature — Bear", 2, 2)
	effectstest.AddCounters(t, g, grown, game.CounterPlusOne, 2)

	id := castSin(t, g)
	effectstest.AnswerEntryChoice(t, g, game.PendingChoiceEntryChoosePermanents, me.ID)
	if got := effectstest.B12Counter(t, g, id, game.CounterPlusOne); got != 0 {
		t.Errorf("Sin has %d counters, want 0", got)
	}
	if got := effectstest.B12Counter(t, g, grown, game.CounterPlusOne); got != 2 {
		t.Errorf("the bear lost counters: %d", got)
	}
}

func TestSinWithNothingCounteredJustEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	id := castSin(t, g)
	if effectstest.EntryChoiceOfKind(g, game.PendingChoiceEntryChoosePermanents, me.ID) != nil {
		t.Error("asked with nothing to remove")
	}
	if _, ok := effectstest.BattlefieldCard(g, id); !ok {
		t.Error("Sin did not enter")
	}
}

// The counters Sin enters with are "enters with" counters, so a
// Doubling Season doubles them (CR 614.1c), and X counts the counters
// that came off, not a doubled removal.
func TestSinEntryCountersAreDoubledByDoublingSeason(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.SeedReplacementPermanent(g, effectstest.DoublingSeasonOracle, "Doubling Season", me.ID)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AddCounters(t, g, theirs, game.CounterPlusOne, 2)

	id := castSin(t, g)
	effectstest.AnswerEntryChoice(t, g, game.PendingChoiceEntryChoosePermanents, me.ID, theirs)
	if got := effectstest.B12Counter(t, g, id, game.CounterPlusOne); got != 8 {
		t.Errorf("Sin entered with %d, want 8 (2 removed, x2, doubled)", got)
	}
}

// The open prompt is a paused entry. Like its ADR 0098 siblings it is
// not a restore point (the census counts its resume frame), but the
// undo clone carries it, and answering on the clone leaves the live
// game alone.
func TestSinPromptSurvivesAnUndoClone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AddCounters(t, g, theirs, game.CounterPlusOne, 2)
	id := castSin(t, g)

	restored := g.Clone()
	c := effectstest.EntryChoiceOfKind(restored, game.PendingChoiceEntryChoosePermanents, me.ID)
	if c == nil {
		t.Fatal("the restored game lost the prompt")
	}
	if len(c.ChooseSuggested) != 1 || c.ChooseSuggested[0] != theirs {
		t.Errorf("restored suggestion %v", c.ChooseSuggested)
	}
	effectstest.AnswerEntryChoice(t, restored, game.PendingChoiceEntryChoosePermanents, me.ID, theirs)
	if got := effectstest.B12Counter(t, restored, id, game.CounterPlusOne); got != 4 {
		t.Errorf("restored Sin entered with %d, want 4", got)
	}
	if got := effectstest.B12Counter(t, g, theirs, game.CounterPlusOne); got != 2 {
		t.Errorf("answering on the clone stripped the live game's bear: %d", got)
	}
	if effectstest.EntryChoiceOfKind(g, game.PendingChoiceEntryChoosePermanents, me.ID) == nil {
		t.Error("answering on the clone closed the live game's prompt")
	}
}

func TestSinDiesMovesItsCountersAndShufflesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sin := effectstest.B12Push(g, me.ID, "Sin, Unending Cataclysm", "Legendary Creature — Leviathan Avatar", sinUnendingCataclysmOracle, 5, 5)
	effectstest.AddCounters(t, g, sin, game.CounterPlusOne, 6)
	effectstest.AddCounters(t, g, sin, "shield", 1)
	keeper := effectstest.B12Creature(g, me.ID, "Keeper", "Creature — Bear", 2, 2)

	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(sin) })
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, keeper)
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.B12Counter(t, g, keeper, game.CounterPlusOne); got != 6 {
		t.Errorf("+1/+1 counters moved: %d, want 6", got)
	}
	if got := effectstest.B12Counter(t, g, keeper, "shield"); got != 1 {
		t.Errorf("every kind moves: shield %d, want 1", got)
	}
	if me.Graveyard.Contains(sin) {
		t.Error("Sin is still in the graveyard")
	}
	if !me.Library.Contains(sin) {
		t.Error("Sin was not shuffled into its owner's library")
	}
}
