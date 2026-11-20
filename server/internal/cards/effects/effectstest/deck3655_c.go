package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// deck3655_c_test.go — The Falcon, Rejoin the Fight and Coin of Fate
// from the Terra, Herald of Hope deck request (#3655).

const (
	Deck3655FalconOracle = "9f4093a6-83f1-4c1d-8b03-8fe15647fb11"
	Deck3655RejoinOracle = "d22ea32b-2128-4dcc-8b1c-c4c52f0b2f45"
	deck3655CoinOracle   = "e43b478b-4cb8-42b1-a5b0-0b1421b84852"
)

// coinOfFateActivation exiles two creature cards from `me`'s graveyard and
// resolves Coin of Fate's ability, leaving whatever prompt it queues open.
func CoinOfFateActivation(t *testing.T, g *game.Game) (me *game.Player, one, two uuid.UUID) {
	t.Helper()
	me = g.Seats[0]
	coin := PushCatalogPermanent(g, me.ID, "Coin of Fate", "Artifact", deck3655CoinOracle, false)
	one = Deck3655GraveyardCreature(me, "One", 1)
	two = Deck3655GraveyardCreature(me, "Two", 1)
	for _, id := range []uuid.UUID{one, two} {
		if err := g.MoveCardByID(
			game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID},
			game.ZoneRef{Kind: game.ZoneExile}, id,
		); err != nil {
			t.Fatalf("exile: %v", err)
		}
	}
	spec, _ := effects.Lookup(deck3655CoinOracle)
	item := &game.StackItem{Kind: game.StackItemActivated, Controller: me.ID, SourceCardID: coin,
		Paid: game.PaidCost{Exiled: []uuid.UUID{one, two}}}
	if err := spec.Activated[0].Effect(g, item); err != nil {
		t.Fatalf("effect: %v", err)
	}
	return me, one, two
}

// finishCoinOfFate checks the cards were split as `one` chosen: bottom of
// the library, `two` back tapped, and the crown taken.
func FinishCoinOfFate(t *testing.T, g *game.Game, me *game.Player, libBefore int) {
	t.Helper()
	if len(me.Library.Cards) != libBefore+1 {
		t.Errorf("library grew by %d, want 1", len(me.Library.Cards)-libBefore)
	}
	var returnedTapped bool
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Two" {
			returnedTapped = c.Tapped
		}
	}
	if !returnedTapped {
		t.Error("the other exiled card did not return tapped")
	}
	if g.Monarch != me.ID {
		t.Error("the controller did not become the monarch")
	}
}
