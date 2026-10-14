package heuristic_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// exile_cast_cost_test.go — #3536: a cast whose additional cost exiles
// graveyard cards (Abhorrent Oculus, exile_ids on cast_spell) spends real
// cards, and the policy has to see them. Before ExileIDs reached
// castParams the payment was invisible and two casts that differed only in
// what they ate priced exactly alike.

func exileCastMove(t *testing.T, seat int, id, label string, exileIDs ...string) legal.Move {
	return legal.Move{
		Type: legal.TypeCastSpell, Player: seatID(seat), Kind: legal.KindCast,
		Label: label, Source: uuid.MustParse(id),
		Params: mustJSON(t, map[string]any{
			"instance_id": id, "from_zone": "hand", "exile_ids": exileIDs,
		}),
	}
}

// Two casts of the same spell that differ only in the graveyard cards they
// exile: one eats a land, the other a creature the seat could escape. The
// policy takes the one that costs it less.
func TestExiledGraveyardCardsArePricedOnACast(t *testing.T) {
	const (
		idDead = "10000000-0000-0000-0000-000000000021"
		idLand = "10000000-0000-0000-0000-000000000022"
	)
	oculus := spell(cardID(20), 0, "Abhorrent Oculus", "{2}{U}")
	v := newView([]protocol.PlayerView{
		newSeat(0, withHand(oculus), withGraveyard(
			graveyardCardView(idLand, 0, "Island", "Basic Land — Island", ""),
			graveyardCardView(idDead, 0, "Uro", "Legendary Creature — Giant Horror", "{1}{G}{U}", castable()),
		)),
		newSeat(1),
	}, withBattlefield(land(cardID(1), 0), land(cardID(2), 0), land(cardID(3), 0)))

	const cheap, dear = "Cast Abhorrent Oculus (exiling Island)", "Cast Abhorrent Oculus (exiling Uro)"
	in := input(0, v,
		passMove(0),
		exileCastMove(t, 0, cardID(20), dear, idDead),
		exileCastMove(t, 0, cardID(20), cheap, idLand),
	)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != cheap {
		t.Errorf("the bot chose %q — the cast that exiles the land should price below the one that exiles a castable creature", got)
	}
}
