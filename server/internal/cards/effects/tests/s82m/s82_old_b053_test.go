package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b053_test.go — Strixhaven Mystical Archive, batch 53 (#5503).

const (
	oldB053StrategicPlanning = "02b5acf3-47cb-4d39-9307-e02656f1879b"
	oldB053GiftOfEstates     = "1bd70be6-752c-4ecb-a3e8-2bacec4b94fc"
	oldB053DivineGambit      = "25fcf381-6f0a-49ef-9c41-f202be7d8368"
)

func TestOldB053EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB053StrategicPlanning: "Strategic Planning",
		oldB053GiftOfEstates:     "Gift of Estates",
		oldB053DivineGambit:      "Divine Gambit",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Completeness != effects.CompletenessFull {
			t.Errorf("%s is %v, want Full", name, spec.Completeness)
		}
	}
}

func TestOldB053StrategicPlanningTakesOneOfThreeBinsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := effectstest.SeedLibrary(me, "One", "Two", "Three", "Four")
	effectstest.CastCatalogSpell(t, g, "Strategic Planning", "Sorcery", oldB053StrategicPlanning, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick among the top three cards")
	}
	if len(pick.ChooseCards) != 3 || pick.ChooseMin != 1 || pick.ChooseMax != 1 {
		t.Fatalf("offered %d cards in [%d,%d], want exactly one of the top three", len(pick.ChooseCards), pick.ChooseMin, pick.ChooseMax)
	}
	for _, id := range pick.ChooseCards {
		if id == ids[3] {
			t.Error("the fourth card was offered")
		}
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{ids[1]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(ids[1]) {
		t.Error("the chosen card is not in hand")
	}
	for _, id := range []uuid.UUID{ids[0], ids[2]} {
		if !me.Graveyard.Contains(id) {
			t.Error("a card that was not chosen is not in the graveyard")
		}
	}
	if !me.Library.Contains(ids[3]) {
		t.Error("the fourth card left the library")
	}
}

func oldB053Plains(me *game.Player, n int) []uuid.UUID {
	var out []uuid.UUID
	for i := 0; i < n; i++ {
		out = append(out, effectstest.PushLibraryCardForTest(me, game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"}))
	}
	return out
}

func TestOldB053GiftOfEstatesFetchesUpToThreePlainsWhenAnOpponentIsAhead(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	plains := oldB053Plains(me, 4)
	// A nonbasic Plains card counts: "Plains cards", not "basic Plains".
	nonbasic := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Kjeldoran Outpost", TypeLine: "Land — Plains"})
	isl := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
	effectstest.PushLand(g, opp.ID, "Mountain")
	effectstest.CastCatalogSpell(t, g, "Gift of Estates", "Sorcery", oldB053GiftOfEstates, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)
	effectstest.PassPriorityAroundTable(t, g)

	got := 0
	for _, id := range append(plains, nonbasic) {
		if effectstest.InHandOf(me, id) {
			got++
		}
	}
	if got != 3 {
		t.Errorf("%d Plains cards in hand, want 3", got)
	}
	if effectstest.InHandOf(me, isl) {
		t.Error("fetched a non-Plains card")
	}
}

func TestOldB053GiftOfEstatesDoesNothingWhenNoOpponentIsAhead(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	plains := oldB053Plains(me, 2)
	// Level: one land each is not "more lands than you".
	effectstest.PushLand(g, me.ID, "Plains")
	effectstest.PushLand(g, opp.ID, "Mountain")
	effectstest.CastCatalogSpell(t, g, "Gift of Estates", "Sorcery", oldB053GiftOfEstates, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SearchPromptFor(g, me.ID) != nil {
		t.Fatal("searched while no opponent controlled more lands")
	}
	for _, id := range plains {
		if effectstest.InHandOf(me, id) {
			t.Error("a Plains card reached the hand")
		}
	}
}

func TestOldB053DivineGambitExilesAndLetsThatPlayerPutAPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	victim := effectstest.PushCatalogPermanent(g, opp.ID, "Bear", "Creature — Bear", "", false)
	perm := effectstest.HandCardForTest(opp, "Angel", "Creature — Angel", "")
	spell := effectstest.HandCardForTest(opp, "Shock", "Instant", "")
	effectstest.CastCatalogSpell(t, g, "Divine Gambit", "Sorcery", oldB053DivineGambit, []game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	effectstest.PassPriorityAroundTable(t, g)

	if effects.OnBattlefield(g, victim) {
		t.Fatal("the target was not exiled")
	}
	pick := effectstest.ChooseCardsChoiceFor(g, opp.ID)
	if pick == nil {
		t.Fatal("the exiled permanent's controller was not offered a permanent card")
	}
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("the caster was asked instead of that player")
	}
	for _, id := range pick.ChooseCards {
		if id == spell {
			t.Error("a nonpermanent card was offered")
		}
	}
	effectstest.AnswerChooseCards(t, g, opp.ID, perm)
	if !effects.OnBattlefield(g, perm) {
		t.Error("the chosen permanent did not enter")
	}
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == perm && c.Controller != opp.ID {
			t.Error("it entered under the wrong player's control")
		}
	}
}

func TestOldB053DivineGambitRefusesYourOwnPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mine := effectstest.PushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	err := effectstest.CastCatalogSpellErr(t, g, "Divine Gambit", "Sorcery", oldB053DivineGambit, []game.TargetRef{{Kind: game.TargetCard, ID: mine}})
	if err == nil {
		t.Error("targeted a permanent its caster controls")
	}
}
