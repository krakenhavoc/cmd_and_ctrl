package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Geyser Drake — Creature — Drake {2}{U}, 2/3 (Outlaws of Thunder
// Junction, #4468):
//
//	"Flying
//	 During turns other than yours, spells you cast cost {1} less to
//	 cast."
//
// A cost modifier on your own spells, generic mana only (CR 601.2f),
// that applies only while another player is the active player. On your
// own turn, and to an opponent's spell on theirs, it does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "411b9158-e427-4099-bd30-3fbb22edb553",
		Name:            "Geyser Drake",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		CostModifiers: []game.CostModifier{
			CostsLess(1, "During turns other than yours, spells you cast cost {1} less to cast.",
				YourSpell(), geyserDrakeNotYourTurn),
		},
	})
}

// geyserDrakeNotYourTurn is "during turns other than yours": the caster
// is not the active player.
func geyserDrakeNotYourTurn(q game.CostQuery) bool {
	return q.Game != nil && !IsYourTurn(q.Game, q.Controller)
}
