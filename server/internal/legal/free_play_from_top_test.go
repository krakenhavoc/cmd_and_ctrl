package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// #3801: "look at the top card of your library, you may cast it free, else
// it goes to your hand" (Ziatora's Envoy). Since #4250 the cast is made
// during the resolution that looks (CR 608.2g): a resolution_cast over the
// looked-at card, still on top of the library. The bot answers it like any
// other: the enumerator offers the library-top cast, naming the prompt, and
// the always-legal decline, and the actions layer accepts each; a declined
// card is in the hand with nothing left to offer.

const oracleTopLooker = "test-4250-top-looker"

func init() {
	effects.Register(effects.Spec{
		OracleID: oracleTopLooker,
		Name:     "Test Top Looker",
		Activated: []effects.ActivatedAbility{{
			Label: "{T}: Look at the top card of your library. You may cast a spell with mana value 3 or less from the top of your library without paying its mana cost. If you don't, put that card into your hand.",
			Cost:  effects.TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return g.OfferFreePlayFromTopOrHandForEffect(game.FreePlayFromTop{
					Player: item.Controller, Source: item.SourceCardID, SourceName: "Test Top Looker", MaxManaValue: 3,
				})
			},
		}},
	})
}

func offerTopSpell(t *testing.T) (g *game.Game, seat *game.Player, top game.Card) {
	t.Helper()
	g = newTable(t)
	seat = g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	seat.Library.Cards = nil
	top = game.NewCard("Free Sorcery", seat.ID)
	top.TypeLine = "Sorcery"
	top.ManaCost = "{1}{U}"
	top.Layout = "normal"
	seat.Library.PushTop(top)
	looker := uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: looker, Name: "Test Top Looker", OracleID: oracleTopLooker,
		TypeLine: "Artifact", Owner: seat.ID, Controller: seat.ID})
	var activate *legal.Move
	for _, m := range legal.EnumerateFor(g, seat.ID) {
		if m.Kind == legal.KindActivate && m.Source == looker {
			activate = &m
			break
		}
	}
	if activate == nil {
		t.Fatal("the looker's activation is not offered")
	}
	if err := actions.Dispatch(g, actions.Action{Type: actions.Type(activate.Type), Player: seat.ID, Caller: seat.ID, Params: activate.Params}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	for i := 0; i < 8 && topOffer(g, seat.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if topOffer(g, seat.ID) == nil {
		t.Fatal("the looker never offered the cast")
	}
	return g, seat, top
}

func topOffer(g *game.Game, seat uuid.UUID) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceResolutionCast && c.Chooser == seat {
			return c
		}
	}
	return nil
}

func TestEnumeratorOffersAFreeCastFromTheLookedAtTop(t *testing.T) {
	g, seat, top := offerTopSpell(t)
	offer := topOffer(g, seat.ID)
	if !g.ResolutionPaused() {
		t.Fatal("the offer is not holding the resolution open")
	}

	moves := legal.EnumerateFor(g, seat.ID)
	got := castMovesFor(moves, top.InstanceID)
	if len(got) == 0 {
		t.Fatalf("no free cast offered off the looked-at top: %v", labels(moves))
	}
	for _, m := range moves {
		if m.ResolutionCast != offer.ID.String() {
			t.Errorf("a seat owing the free cast was offered %q as well", m.Label)
		}
	}
	dispatchAll(t, g, seat.ID, moves)

	// The decline puts the card into the hand, with nothing left to
	// offer.
	if err := g.ResolvePendingChoice(offer.ID, seat.ID, nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	if !seat.Hand.Contains(top.InstanceID) || g.ResolutionPaused() {
		t.Fatal("the declined card is not in the hand, or the resolution is still held")
	}
	if got := castMovesFor(legal.EnumerateFor(g, seat.ID), top.InstanceID); len(got) != 0 {
		t.Errorf("a declined card is still offered as a free cast: %v", labels(got))
	}
}
