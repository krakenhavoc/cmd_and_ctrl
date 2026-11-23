package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scalding Devil — {1}{R} Creature — Devil 1/1 (Avacyn Restored, #5570):
//
//	"{2}{R}: This creature deals 1 damage to target player or
//	 planeswalker."
//
// Goblin Fireslinger's ping at a player or planeswalker, paid with
// {2}{R} instead of a tap, so it can be activated more than once a
// turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3bb1536f-06cc-433f-aeeb-afd528f9af03",
		Name:         "Scalding Devil",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R}: This creature deals 1 damage to target player or planeswalker.",
			Cost:    effects.ManaCost("{2}{R}"),
			Targets: effects.TargetPlayerOrPlaneswalker(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Effect:  effects.SourceDealsOneToFirstTarget,
		}},
	})
}
