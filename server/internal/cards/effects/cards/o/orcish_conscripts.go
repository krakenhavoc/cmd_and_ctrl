package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Orcish Conscripts — Creature — Orc {R}, 2/2 (Ice Age, S82 OLD, #5636):
//
//	"This creature can't attack unless at least two other creatures attack.
//	 This creature can't block unless at least two other creatures block."
//
// Two restrictions checked on the whole declaration (CR 508.1c, 509.1b): an attack with fewer than two
// other attackers, or a block with fewer than two other blockers of the same player, is illegal.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d99e1d08-32ca-49c5-97cc-e064d661202f",
		Name:         "Orcish Conscripts",
		Completeness: effects.CompletenessFull,
		AttackLimits: []game.AttackLimit{effects.CantAttackUnlessNOtherCreaturesAttack(2)},
		BlockRules:   []game.BlockRule{effects.CantBlockUnlessNOtherCreaturesBlock(2)},
	})
}
