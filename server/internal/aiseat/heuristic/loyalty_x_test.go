package heuristic_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// loyalty_x_test.go — #1944: the bot's policy for a −X loyalty cost.
// The enumerator offers every X the loyalty pays, each with its own
// loyalty price; a damage row declared DamageIsX is priced at the X
// the move names, so the smallest lethal X wins: a smaller one kills
// nothing, and a bigger one spends loyalty for nothing.

func minusXMove(t *testing.T, src, target string, x int) legal.Move {
	t.Helper()
	return legal.Move{
		Type: legal.TypeActivateAbility, Player: seatID(0), Kind: legal.KindActivate,
		Label:  fmt.Sprintf("Chandra: −X for X=%d", x),
		Source: uuid.MustParse(src),
		Cost:   &legal.MoveCost{Loyalty: -x},
		Params: mustJSON(t, map[string]any{
			"source_card_id": src, "ability_index": 0, "x_value": x,
			"targets": []map[string]any{{"kind": "card", "id": target}},
		}),
	}
}

func TestMinusXPicksTheSmallestLethalX(t *testing.T) {
	zero := 0
	chandra := walkerView(cardID(1), 0, 6)
	chandra.ActivatedAbilities = []protocol.ActivatedAbilityView{{
		Index: 0, Ref: "own:0", Label: "−X: Chandra deals X damage to target creature or planeswalker.",
		LoyaltyCost: &zero, LoyaltyCostX: true, DemandsX: true, SorcerySpeed: true,
		Purpose: targetEntries(protocol.TargetPurposeView{Slot: 0, DamageIsX: true}),
	}}
	ogre := creature(cardID(2), 1, "Ogre", 3, 3)
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(chandra, ogre), withTurn(3, 0, "precombat_main"))
	moves := []legal.Move{passMove(0)}
	for x := 1; x <= 6; x++ {
		moves = append(moves, minusXMove(t, chandra.InstanceID, ogre.InstanceID, x))
	}
	in := input(0, v, moves...)
	best, bestV := -1, 0.0
	for _, c := range heuristic.New().Rank(context.Background(), in) {
		if best < 0 || c.Value > bestV {
			best, bestV = c.Index, c.Value
		}
	}
	if got := in.Moves[best].Label; got != "Chandra: −X for X=3" {
		t.Errorf("best move = %q, want X=3 (the 3/3's toughness)", got)
	}
}
