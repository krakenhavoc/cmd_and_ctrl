package aiseat_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// resolution_door_cast_bot_test.go — #4250 through the real engine, the
// real enumerator and the real heuristic. Rebound's, madness's and the
// free-play-from-top offer's casts are resolution_cast prompts answered
// inside the resolution that offers them (CR 608.2g): the enumerator
// offers the casts and the always-legal decline, every move names the
// prompt, the heuristic must answer (a seat that owes a choice and never
// takes a move wedges the table, #544), and the engine accepts what it
// chose. Not behind AISEAT_GAME_TESTS: a few decisions each.

const botFieryTemperOracle = "f07bd49d-8e71-4d56-be2a-638514011318"

const botTopLookerOracle = "test-4250-bot-top-looker"

func init() {
	effects.Register(effects.Spec{
		OracleID: botTopLookerOracle,
		Name:     "Test Bot Top Looker",
		Activated: []effects.ActivatedAbility{{
			Label: "{T}: Look at the top card of your library. You may cast a spell with mana value 3 or less from the top of your library without paying its mana cost. If you don't, put that card into your hand.",
			Cost:  effects.TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return g.OfferFreePlayFromTopOrHandForEffect(game.FreePlayFromTop{
					Player: item.Controller, Source: item.SourceCardID, SourceName: "Test Bot Top Looker", MaxManaValue: 3,
				})
			},
		}},
	})
}

// botAnswersTheDoor checks the enumerated answers to me's open
// resolution_cast and lets the heuristic take one. It returns the card
// on offer.
func botAnswersTheDoor(t *testing.T, g *game.Game, me uuid.UUID, keyword string) uuid.UUID {
	t.Helper()
	var offer *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceResolutionCast && c.Chooser == me {
			offer = c
		}
	}
	if offer == nil || offer.MayCastKeyword != keyword || len(offer.DiscardOptions) != 1 {
		t.Fatalf("offer = %+v, want a %s resolution_cast over one card", offer, keyword)
	}
	if !g.ResolutionPaused() {
		t.Fatal("the offer is not inside the resolution that made it")
	}
	card := offer.DiscardOptions[0]
	var casts, declines int
	for _, mv := range legal.EnumerateFor(g, me) {
		switch {
		case mv.ResolutionCast != offer.ID.String():
			t.Errorf("move %q does not answer the %s offer", mv.Label, keyword)
		case mv.Kind == legal.KindCast && mv.Source == card:
			casts++
		case mv.Kind == legal.KindChoice && mv.AlwaysLegal:
			declines++
		}
	}
	if casts == 0 || declines != 1 {
		t.Fatalf("offered %d casts and %d declines", casts, declines)
	}
	taken := driveChoices(t, g, heuristic.New(), me, 4)
	if len(taken) != 1 {
		t.Fatalf("answered the %s offer with %v", keyword, taken)
	}
	if g.ResolutionPaused() {
		t.Errorf("the resolution is still held after the bot's answer %v", taken)
	}
	return card
}

// Rebound: the upkeep trigger's free cast of a sorcery. Free and harmless,
// so the bot casts it.
func TestHeuristicSeatAnswersAReboundResolutionCast(t *testing.T) {
	g := newSettledTable(t, 4250)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToStep(t, g, game.StepPrecombatMain)
	me.Hand.Cards = nil
	me.Command.Cards = nil
	id := uuid.New()
	c := game.Card{InstanceID: id, Name: "Test Rebound Sorcery", TypeLine: "Sorcery", ManaCost: "{1}{U}",
		Layout: "normal", Keywords: []string{game.KeywordRebound}, Owner: me.ID, Controller: me.ID}
	c.AddKnowersAll(seatIDs(g))
	me.Hand.PushTop(c)
	for i := 0; i < 2; i++ {
		g.Battlefield.PushTop(game.Card{InstanceID: uuid.New(), Name: "Island", TypeLine: "Basic Land — Island",
			Owner: me.ID, Controller: me.ID})
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the rebound sorcery: %v", err)
	}
	for i := 0; i < 16 && !g.Exile.Contains(id); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !g.Exile.Contains(id) {
		t.Fatal("setup: the rebound sorcery was not exiled")
	}
	seat := g.Turn.ActiveSeat
	for i := 0; i < 80; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
		if g.Turn.Step == game.StepUpkeep && g.Turn.ActiveSeat == seat {
			break
		}
	}
	for i := 0; i < 16 && !owesChoice(g, me.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !owesChoice(g, me.ID) {
		t.Fatal("rebound asked nothing at the upkeep")
	}
	botAnswersTheDoor(t, g, me.ID, game.MayCastKeywordRebound)
	if !g.Stack.Contains(id) {
		t.Error("the bot did not take the free rebound cast")
	}
}

// Madness: a discarded Fiery Temper offered for {R}, with a Mountain to
// pay. Whatever the bot chooses, it answers, and the engine accepts it.
func TestHeuristicSeatAnswersAMadnessResolutionCast(t *testing.T) {
	g := newSettledTable(t, 42501)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToStep(t, g, game.StepPrecombatMain)
	me.Hand.Cards = nil
	me.Command.Cards = nil
	temper := uuid.New()
	c := game.Card{InstanceID: temper, Name: "Fiery Temper", OracleID: botFieryTemperOracle, TypeLine: "Instant",
		ManaCost: "{1}{R}{R}", Colors: []string{"R"}, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(c)
	g.Battlefield.PushTop(game.Card{InstanceID: uuid.New(), Name: "Mountain", TypeLine: "Basic Land — Mountain",
		Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() {
		if err := g.DiscardRandomForEffect(me.ID, 1); err != nil {
			t.Fatalf("discard: %v", err)
		}
	})
	if !g.Exile.Contains(temper) {
		t.Fatal("setup: madness did not exile the discarded card")
	}
	for i := 0; i < 16 && !owesChoice(g, me.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !owesChoice(g, me.ID) {
		t.Fatal("madness asked nothing")
	}
	card := botAnswersTheDoor(t, g, me.ID, game.MayCastKeywordMadness)
	if !g.Stack.Contains(card) && !me.Graveyard.Contains(card) {
		t.Error("the madness card was neither cast nor put into the graveyard")
	}
}

// The free play from the top of the library: a free two-drop sorcery
// looked at by an activated ability. The bot casts it from the library.
func TestHeuristicSeatAnswersAFreeCastFromTheTop(t *testing.T) {
	g := newSettledTable(t, 42502)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToStep(t, g, game.StepPrecombatMain)
	me.Hand.Cards = nil
	me.Command.Cards = nil
	top := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: top, Name: "Free Sorcery", TypeLine: "Sorcery", ManaCost: "{1}{U}",
		Layout: "normal", Owner: me.ID, Controller: me.ID})
	looker := uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: looker, Name: "Test Bot Top Looker", OracleID: botTopLookerOracle,
		TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	if err := g.ActivateCatalogAbility(me.ID, looker, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	for i := 0; i < 16 && !owesChoice(g, me.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !owesChoice(g, me.ID) {
		t.Fatal("the looker asked nothing")
	}
	botAnswersTheDoor(t, g, me.ID, game.MayCastKeywordFreeOrHand)
	if !g.Stack.Contains(top) {
		t.Error("the bot did not cast the free spell from the top of its library")
	}
}
