package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Warded Battlements — Creature — Wall {2}{W}, 0/3 (Core Set 2021, #5511):
//
//	"Defender
//	 Attacking creatures you control get +1/+0."
//
// Goblin Oriflamme's static (attackingCreaturesYouControlGetPower) on a
// defender. It reads attacking creatures as the layers are applied, so a
// creature gets the bonus only while it is attacking.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7616dacf-44e2-4b03-b120-875cc2e50ea5",
		Name:            "Warded Battlements",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Static:          []game.StaticAbility{effects.AttackingCreaturesYouControlGetPower(1)},
	})
}
