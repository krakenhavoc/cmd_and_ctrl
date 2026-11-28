package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Heartwood Dryad — Creature — Dryad {1}{G}, 2/1 (Tempest, S82 OLD, #5628):
//
//	"This creature can block creatures with shadow as though it had shadow."
//
// Dragon Hunter's waiver for shadow: the one restriction CR 702.28b puts on a
// blocker without shadow facing an attacker that has it. The Dryad still blocks
// every creature without shadow as any creature does, and does not gain shadow
// (nothing that counts creatures with shadow sees it).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "976968c5-7ef4-4451-86ff-03a011bbcca7",
		Name:         "Heartwood Dryad",
		Completeness: effects.CompletenessFull,
		BlockRules:   []game.BlockRule{effects.CanBlockShadowAsThoughItHadShadow()},
	})
}
