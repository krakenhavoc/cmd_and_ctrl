package heuristic_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// exile_cost_test.go — #1297: an activation whose cost exiles cards
// ("Exile a card from your hand", "Exile two cards from your graveyard")
// spends real resources, and the policy has to see them. Before the
// ExileIDs field reached activateParams the payment was invisible and
// an exile-cost activation priced exactly like a free one.

func exileActivateMove(t *testing.T, seat int, src, label string, exileIDs ...string) legal.Move {
	params := map[string]any{"source_card_id": src, "ability_index": 0}
	if len(exileIDs) > 0 {
		params["exile_ids"] = exileIDs
	}
	return legal.Move{
		Type: legal.TypeActivateAbility, Player: seatID(seat), Kind: legal.KindActivate,
		Label: label, Source: uuid.MustParse(src),
		Params: mustJSON(t, params),
	}
}

// Two activations that differ only in what they exile: one pays with a
// spell the seat could cast next turn, the other pays nothing. The
// policy takes the one that costs it nothing.
func TestExiledCardsArePricedOnAnActivation(t *testing.T) {
	held := spell(cardID(20), 0, "Held Spell", "{2}{G}")
	v := newView([]protocol.PlayerView{newSeat(0, withHand(held)), newSeat(1)},
		withBattlefield(
			land(cardID(1), 0),
			land(cardID(2), 0),
		))
	const costly, free = "Pays A Card: exile a card from your hand", "Pays Nothing: do the same thing"
	in := input(0, v,
		passMove(0),
		exileActivateMove(t, 0, cardID(1), costly, cardID(20)),
		exileActivateMove(t, 0, cardID(2), free),
	)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != free {
		t.Errorf("the bot chose %q — the activation that exiles a card from hand should price below the free one", got)
	}
}

// ADR 0137 (#2124): a craft material may be a card in the seat's own
// graveyard, named in exile_permanent_ids beside the permanents. It is
// fuel, and priced as fuel: an activation that spends one costs more
// than the same activation that spends nothing. Before craft, an id in
// that list that was not on the battlefield was priced at zero.
func TestAGraveyardCraftMaterialIsPricedAsFuel(t *testing.T) {
	dead := graveyardCardView(cardID(30), 0, "Dead Bear", "Creature — Bear", "{1}{G}")
	v := newView([]protocol.PlayerView{newSeat(0, withGraveyard(dead)), newSeat(1)},
		withBattlefield(
			land(cardID(1), 0),
			land(cardID(2), 0),
		))
	const costly, free = "Eats A Graveyard Card: craft", "Eats Nothing: the same ability"
	pays := legal.Move{
		Type: legal.TypeActivateAbility, Player: seatID(0), Kind: legal.KindActivate,
		Label: costly, Source: uuid.MustParse(cardID(1)),
		Params: mustJSON(t, map[string]any{
			"source_card_id": cardID(1), "ability_index": 0,
			"exile_permanent_ids": []string{cardID(30)},
		}),
	}
	in := input(0, v, passMove(0), pays, exileActivateMove(t, 0, cardID(2), free))
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != free {
		t.Errorf("the bot chose %q — exiling a graveyard card to a cost should price below paying nothing", got)
	}
}
