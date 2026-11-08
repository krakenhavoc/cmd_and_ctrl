package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Highspire Bell-Ringer — Creature — Djinn Monk {2}{U}, 1/4 (Tarkir:
// Dragonstorm, #4463):
//
//	"Flying
//	 The second spell you cast each turn costs {1} less to cast."
//
// Alisaie Leveilleur's discount for {1}: it reads the per-turn cast tally
// (secondSpellYouCastThisTurn), where the spell being priced is not yet
// counted, and takes {1} off the generic part of the cost only (CR
// 601.2f). The first and the third spell pay in full.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6d2d250b-cdce-4075-8724-b52ad31664fb",
		Name:            "Highspire Bell-Ringer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		CostModifiers: []game.CostModifier{
			CostsLess(1, "The second spell you cast each turn costs {1} less to cast.",
				YourSpell(), secondSpellYouCastThisTurn()),
		},
	})
}
