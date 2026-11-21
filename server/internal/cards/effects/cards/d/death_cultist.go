package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Death Cultist — Creature — Human Wizard {B}, 1/1 (Rise of the Eldrazi,
// #5581):
//
//	"Sacrifice this creature: Target player loses 1 life and you gain 1
//	 life."
//
// The drain is the one Gnawing Zombie's ability makes; with its target gone
// by resolution the ability does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "edbde032-4432-4610-83a2-a52f6ce199e1",
		Name:         "Death Cultist",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this creature: Target player loses 1 life and you gain 1 life.",
			Cost:    effects.SacrificeThis(),
			Targets: effects.TargetPlayer("target player"),
			Effect:  effects.DrainTargetOpponentOne,
		}},
	})
}
