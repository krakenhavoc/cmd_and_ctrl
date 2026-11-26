package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tidal Force — Creature — Elemental {5}{U}{U}{U}, 7/7 (Commander 2013, S82
// older sets):
//
//	"At the beginning of each upkeep, you may tap or untap target permanent."
//
// Bounding Krasis's shape: the target is chosen as the trigger goes on the
// stack (at every player's upkeep, Tidal Force's controller choosing), judged
// again as it resolves (CR 608.2b), and the tap or untap, or neither, is the
// shared "may tap or untap" prompt.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1b25e262-e2df-4768-b55e-1b7b8d3ee993",
		Name:         "Tidal Force",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.AtEachUpkeep("Tidal Force — you may tap or untap target permanent", effects.NemMayTapOrUntapTheTarget("Tidal Force")),
				effects.TargetPermanent("target permanent")),
		},
	})
}
