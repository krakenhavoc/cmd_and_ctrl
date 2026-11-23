package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Battering Wurm — Creature — Wurm {6}{G}, 4/3 (Guildpact, #5599):
//
//	"Bloodthirst 1 (If an opponent was dealt damage this turn, this
//	 creature enters with a +1/+1 counter on it.)
//	 Creatures with power less than this creature's power can't block it."
//
// Bloodthirst is its own entry replacement (Bloodthirst). The block rule
// is Howlgeist's: the comparison is made as blockers are declared,
// against the Wurm's power then, so the bloodthirst counter raises the
// bar.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fdcfa49a-70ed-48a4-b424-527a1698fd85",
		Name:         "Battering Wurm",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.Bloodthirst(1, "Battering Wurm")},
		BlockRules: []game.BlockRule{
			effects.CantBlockAttackers(effects.PowerLessThanSource(), effects.OnSelf(),
				"creatures with power less than Battering Wurm's can't block it"),
		},
	})
}
