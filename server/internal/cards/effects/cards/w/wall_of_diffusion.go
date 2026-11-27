package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wall of Diffusion — Creature — Wall {1}{R}, 0/5 (Tempest, S82 OLD batch 767,
// #5628):
//
//	"Defender
//	 This creature can block creatures with shadow as though it had shadow."
//
// Dragon Hunter's reach waiver with shadow in its place (BlockRule.ShadowAgainst,
// added with this card): a waiver of CR 702.28b for this creature's own blocks
// of attackers that have shadow, read live. It does not gain shadow, so it
// blocks nothing else new, and it still can't block a creature with flying it
// has no way to reach, or any other evasion. Defender is a printed keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ae42f803-f4ac-45f3-9ce7-227a59096f19",
		Name:            "Wall of Diffusion",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		BlockRules:      []game.BlockRule{effects.CanBlockAsThoughItHadShadow(effects.CreaturesWithShadow())},
	})
}
