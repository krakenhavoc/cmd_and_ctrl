package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Bird-Grabber — Creature — Goblin {1}{R}, 2/1 (Core Set 2020, #5518):
//
//	"{R}: This creature gains flying until end of turn. Activate only if you
//	 control a creature with flying."
//
// Celestial Enforcer's gate (ControlsAtLeast over a creature with flying,
// read through the layers) on a self flying grant.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "780f8be6-d477-43ca-92ba-0fe480bbd54d",
		Name:         "Goblin Bird-Grabber",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     "{R}: This creature gains flying until end of turn. Activate only if you control a creature with flying.",
			Purpose:   game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:      effects.ManaCost("{R}"),
			Condition: effects.ControlsAtLeast(1, effects.BnsB27CreatureWithFlying),
			Effect:    effects.ThisGainsKeywordUntilEndOfTurn("flying", "Goblin Bird-Grabber — flying"),
		}},
	})
}
