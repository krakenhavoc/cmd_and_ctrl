package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Plague Spitter — {2}{B} Creature — Phyrexian Horror 2/2 (Jumpstart 2022, #5948):
//
//	"At the beginning of your upkeep, this creature deals 1 damage to each
//	 creature and each player.
//	 When this creature dies, it deals 1 damage to each creature and each
//	 player."
//
// Pestilence's damage, once as an upkeep trigger and once as a dies
// trigger (CR 700.4); the damage is dealt in one instance by the source.
//
// No simplification.
func init() {
	sweep := game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1}}
	damage := func(g *game.Game, item *game.StackItem) error {
		return effects.B23DamageEachCreatureAndEachPlayer(effects.NewContext(g, item), 1)
	}
	effects.Register(effects.Spec{
		OracleID:     "5feedfb0-30e6-400d-9e28-d541ea1aa14e",
		Name:         "Plague Spitter",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.AtYourUpkeep("Plague Spitter — 1 damage to each creature and each player", damage), sweep),
			effects.TriggerWithPurpose(effects.WhenThisDies("Plague Spitter — 1 damage to each creature and each player", damage), sweep),
		},
	})
}
