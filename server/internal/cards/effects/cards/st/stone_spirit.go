package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stone Spirit — Creature — Elemental Spirit {4}{R}, 4/3 (Ice Age, S82 OLD, #5636):
//
//	"This creature can't be blocked by creatures with flying."
//
// Greater Stone Spirit's block rule on the creature itself (CR 509.1b), read through the layers.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "63bc0a02-0d6c-4a00-8b17-1b024f830219",
		Name:         "Stone Spirit",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedBy(effects.OnSelf(), effects.HasKeyword("flying"), "creatures with flying"),
		},
	})
}
