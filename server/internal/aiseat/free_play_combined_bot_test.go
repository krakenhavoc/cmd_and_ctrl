package aiseat_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// free_play_combined_bot_test.go — #4697 through the real engine, the real
// enumerator and the real heuristic. Ziatora's Envoy's "play a land or cast
// a spell from the top of your library" over a pathway (two land faces) is
// the combined resolution_cast: the enumerator offers a land play of each
// face and the decline, every move names the prompt, and the heuristic
// answers it (a seat that owes a choice and never takes a move wedges the
// table, #544).

const botTopLandLookerOracle = "test-4697-bot-top-land-looker"

func init() {
	effects.Register(effects.Spec{
		OracleID: botTopLandLookerOracle,
		Name:     "Test Bot Top Land Looker",
		Activated: []effects.ActivatedAbility{{
			Label: "{T}: Look at the top card of your library. You may play a land from the top of your library or cast a spell with mana value 3 or less from the top of your library without paying its mana cost. If you don't, put that card into your hand.",
			Cost:  effects.TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return g.OfferFreePlayFromTopOrHandForEffect(game.FreePlayFromTop{
					Player: item.Controller, Source: item.SourceCardID, SourceName: "Test Bot Top Land Looker",
					MaxManaValue: 3, Land: true,
				})
			},
		}},
	})
}

func TestHeuristicSeatAnswersACombinedLandPlayOffer(t *testing.T) {
	g := newSettledTable(t, 46970)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToStep(t, g, game.StepPrecombatMain)
	me.Hand.Cards = nil
	me.Command.Cards = nil
	top := game.Card{
		InstanceID: uuid.New(), Layout: game.LayoutModalDFC, Owner: me.ID, Controller: me.ID,
		Faces: []game.Face{
			{Name: "Test Dawn Pathway", TypeLine: "Land"},
			{Name: "Test Dusk Pathway", TypeLine: "Land"},
		},
	}
	top.SetFace(0)
	me.Library.PushTop(top)
	looker := uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: looker, Name: "Test Bot Top Land Looker", OracleID: botTopLandLookerOracle,
		TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	if err := g.ActivateCatalogAbility(me.ID, looker, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	for i := 0; i < 16 && !owesChoice(g, me.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	var offer *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceResolutionCast && c.Chooser == me.ID {
			offer = c
		}
	}
	if offer == nil || offer.MayCastKeyword != game.MayCastKeywordPlayOrFreeOrHand || !g.ResolutionPaused() {
		t.Fatalf("offer = %+v, want the combined offer inside the resolution", offer)
	}
	var plays, declines int
	for _, mv := range legal.EnumerateFor(g, me.ID) {
		switch {
		case mv.ResolutionCast != offer.ID.String():
			t.Errorf("move %q does not answer the offer", mv.Label)
		case mv.Kind == legal.KindLand && mv.Source == top.InstanceID:
			plays++
		case mv.Kind == legal.KindChoice && mv.AlwaysLegal:
			declines++
		}
	}
	if plays != 2 || declines != 1 {
		t.Fatalf("offered %d land plays and %d declines, want 2 and 1", plays, declines)
	}
	taken := driveChoices(t, g, heuristic.New(), me.ID, 4)
	if len(taken) != 1 {
		t.Fatalf("answered the offer with %v", taken)
	}
	if g.ResolutionPaused() {
		t.Errorf("the resolution is still held after the bot's answer %v", taken)
	}
	onField := false
	for _, c := range g.Battlefield.Cards {
		onField = onField || (c.InstanceID == top.InstanceID && c.IsLand())
	}
	if !onField && !me.Hand.Contains(top.InstanceID) {
		t.Errorf("the pathway is neither on the battlefield nor in the hand after %v", taken)
	}
}
