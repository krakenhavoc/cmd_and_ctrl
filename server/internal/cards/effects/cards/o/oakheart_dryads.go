package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Oakheart Dryads — Enchantment Creature — Nymph Dryad {2}{G}, 2/3 (Journey into Nyx, #5558):
//
//	"Constellation — Whenever this creature or another enchantment you
//	 control enters, target creature gets +1/+1 until end of turn."
//
// Constellation is Eidolon of Blossoms' condition: its own entry (it is an
// enchantment) or another enchantment entering under your control.
// Guardian of Pilgrims' pump on the still-legal target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "60bac6e8-0c72-4665-9b01-d07c99dd1716",
		Name:         "Oakheart Dryads",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventETB, effects.ConstellationTriggers, "Oakheart Dryads — target creature gets +1/+1 until end of turn (constellation)",
				effects.PumpAndGrantFirstTarget(1, 1, "Oakheart Dryads — +1/+1")), effects.TargetCreature("target creature")),
		},
	})
}
