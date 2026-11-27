package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Cohort — Creature — Goblin Warrior {R}, 2/2 (Betrayers of Kamigawa, S82 OLD b595):
//
//	"This creature can't attack unless you've cast a creature spell this turn."
//
// An attack-target restriction (CR 508.1c) on the creature itself, read live at declaration
// from the per-turn cast tally (AttackTargetRestriction.ControllerMustHaveCastCreature,
// Silburlind Snapper's twin).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "88f9a7e3-0b35-43e4-b16a-19c54a1e4e2c",
		Name:         "Goblin Cohort",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.OldB595CantAttackUnlessYouHaveCastACreatureSpell()},
	})
}
