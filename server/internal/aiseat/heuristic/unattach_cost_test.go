package heuristic_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// unattach_cost_test.go — #4799: an activation whose cost unattaches an
// Equipment (Captain America, First Avenger's throw) gives up what the
// Equipment was doing, and the policy has to see it. Without the
// UnattachIDs field on activateParams the payment was invisible and the
// activation priced exactly like a free one.

// Two activations that differ only in what they unattach: one takes an
// Equipment off, the other pays nothing. The policy takes the free one.
func TestAnUnattachedEquipmentIsPricedOnAnActivation(t *testing.T) {
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(
			land(cardID(1), 0),
			land(cardID(2), 0),
			equipment(cardID(30), 0, "Test Hammer"),
		))
	const costly, free = "Unattaches The Hammer: throw it", "Unattaches Nothing: the same ability"
	pays := legal.Move{
		Type: legal.TypeActivateAbility, Player: seatID(0), Kind: legal.KindActivate,
		Label: costly, Source: uuid.MustParse(cardID(1)),
		Params: mustJSON(t, map[string]any{
			"source_card_id": cardID(1), "ability_index": 0,
			"unattach_ids": []string{cardID(30)},
		}),
	}
	in := input(0, v, passMove(0), pays, exileActivateMove(t, 0, cardID(2), free))
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != free {
		t.Errorf("the bot chose %q — unattaching an Equipment to a cost should price below paying nothing", got)
	}
}
