package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// ADR 0152 §6: a cast of another face is priced as that face. Megatron
// shape: the converted price buys a 4/5 Vehicle for 4 where the printed
// price buys a 7/5 creature for 6, and with both affordable the bot must
// not take the converted price for the same card it prices at the Robot.
func twoFacedCast(t *testing.T, converted protocol.CardFaceView) (printed, conv float64) {
	t.Helper()
	sp := spell(cardID(1), 0, "Megatron", "{4}{B}{B}")
	sp.TypeLine = "Legendary Artifact Creature — Robot"
	sp.Power, sp.Toughness = 7, 5
	sp.Layout = "transform"
	sp.Faces = []protocol.CardFaceView{
		{Name: "Megatron", TypeLine: sp.TypeLine, ManaCost: "{4}{B}{B}", Power: 7, Toughness: 5},
		converted,
	}
	sp.AlternativeCosts = []protocol.AlternativeCostView{{
		Key: "more_than_meets_the_eye", Label: "More Than Meets the Eye {2}{B}{B}",
		ManaCost: "{2}{B}{B}", CastsFace: 1,
	}}
	hard := castMove(t, 0, sp.InstanceID, "Cast Megatron")
	hard.Params = mustJSON(t, map[string]any{"instance_id": sp.InstanceID, "from_zone": "hand"})
	cv := castMove(t, 0, sp.InstanceID, "Cast Megatron converted")
	cv.Params = mustJSON(t, map[string]any{"instance_id": sp.InstanceID, "from_zone": "hand",
		"alternative_cost": "more_than_meets_the_eye", "face": 1})
	v := newView([]protocol.PlayerView{newSeat(0, withHand(sp)), newSeat(1)},
		withBattlefield(land(cardID(2), 0)), withTurn(9, 0, "precombat_main"))
	in := input(0, v, passMove(0), hard, cv)
	pol := heuristic.New()
	return rankValue(t, pol, in, hard.Label), rankValue(t, pol, in, cv.Label)
}

func TestConvertedCastIsPricedAsTheFaceItCasts(t *testing.T) {
	printed, conv := twoFacedCast(t, protocol.CardFaceView{
		Name: "Megatron, Tyrant", TypeLine: "Legendary Artifact — Vehicle", Power: 4, Toughness: 5,
		OracleText: "Living metal (During your turn, this Vehicle is also a creature.)",
	})
	if conv >= printed {
		t.Errorf("converted 4/5 Vehicle priced %.3f, printed 7/5 creature %.3f; want the printed cast above", conv, printed)
	}
}

func TestLivingMetalFaceIsPricedAsACreature(t *testing.T) {
	_, withMetal := twoFacedCast(t, protocol.CardFaceView{
		Name: "Megatron, Tyrant", TypeLine: "Legendary Artifact — Vehicle", Power: 4, Toughness: 5,
		OracleText: "Living metal (During your turn, this Vehicle is also a creature.)",
	})
	_, plain := twoFacedCast(t, protocol.CardFaceView{
		Name: "Plain Vehicle", TypeLine: "Legendary Artifact — Vehicle", Power: 4, Toughness: 5,
	})
	if withMetal <= plain {
		t.Errorf("a living-metal face priced %.3f, a plain Vehicle %.3f; want the creature above", withMetal, plain)
	}
}
