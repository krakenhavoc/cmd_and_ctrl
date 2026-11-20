package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "780f8be6-d477-43ca-92ba-0fe480bbd54d",
		Name:         "Goblin Bird-Grabber",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     "{R}: This creature gains flying until end of turn. Activate only if you control a creature with flying.",
			Purpose:   game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:      ManaCost("{R}"),
			Condition: ControlsAtLeast(1, bnsB27CreatureWithFlying),
			Effect:    thisGainsKeywordUntilEndOfTurn("flying", "Goblin Bird-Grabber — flying"),
		}},
	})
}
