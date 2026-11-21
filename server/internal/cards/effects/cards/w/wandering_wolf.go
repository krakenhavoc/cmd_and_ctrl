package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wandering Wolf — {1}{G} Creature — Wolf 2/1 (Avacyn Restored, #5570):
//
//	"Creatures with power less than this creature's power can't block
//	 it."
//
// Howlgeist's block rule: a creature whose power is less than this
// creature's, judged as blockers are declared, can't block it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1750cf0a-0709-4f02-90ea-572ff88a37a5",
		Name:         "Wandering Wolf",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBlockAttackers(effects.PowerLessThanSource(), effects.OnSelf(),
				"creatures with power less than Wandering Wolf's can't block it"),
		},
	})
}
