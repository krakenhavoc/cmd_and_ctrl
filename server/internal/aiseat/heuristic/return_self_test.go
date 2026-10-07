package heuristic_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// #2028: a row that returns its own source to hand pays half of the
// permanent (it comes back and can be cast again) and all of a token
// (which ceases to exist in the hand, CR 111.7), whether or not the row
// declares a purpose.
func TestAReturnThisRowPaysForTheSourceItGivesUp(t *testing.T) {
	pol := heuristic.New()
	price := func(returnSelf, token bool, purpose *protocol.PurposeView) float64 {
		src := rock(cardID(1), 0, "Bouncing Rock", "{1}", protocol.ManaAbilityView{})
		src.ManaAbilities = nil
		src.IsToken = token
		src.ActivatedAbilities = []protocol.ActivatedAbilityView{{
			Index: 0, ReturnSelf: returnSelf, ManaCost: "{2}", Purpose: purpose,
		}}
		act := legal.Move{
			Type: legal.TypeActivateAbility, Player: seatID(0), Kind: legal.KindActivate,
			Label: "Activate Bouncing Rock", Source: uuid.MustParse(src.InstanceID),
			Params: mustJSON(t, map[string]any{"source_card_id": src.InstanceID, "ability_index": 0}),
		}
		bf := append(manaLands(6, 0, 100), src)
		v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)}, withBattlefield(bf...), withTurn(3, 0, "precombat_main"))
		return rankValue(t, pol, input(0, v, passMove(0), act), act.Label)
	}

	// The rock is worth 1.2 on the board (the Bauble's price in
	// TestAnActivatedRowIsPricedByItsDeclaredPurpose).
	for _, purpose := range []*protocol.PurposeView{nil, {Draws: 1}} {
		if kept, returned := price(false, false, purpose), price(true, false, purpose); !nearly(kept-returned, 0.6) {
			t.Errorf("purpose %+v: returning the card cost %.3f, want half of 1.2", purpose, kept-returned)
		}
		if kept, returned := price(false, true, purpose), price(true, true, purpose); !nearly(kept-returned, 1.2) {
			t.Errorf("purpose %+v: returning a token cost %.3f, want all of 1.2", purpose, kept-returned)
		}
	}
}
