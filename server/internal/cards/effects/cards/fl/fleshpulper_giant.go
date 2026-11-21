package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fleshpulper Giant — Creature — Giant {5}{R}{R}, 4/4 (Magic 2014, #5563):
//
//	"When this creature enters, you may destroy target creature with toughness 2 or less."
//
// An optional enters trigger (CR 603.5): the target is chosen as it goes on
// the stack, and the "you may" is asked as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7cdeecb5-354a-4a6c-989f-d8c9554befb7",
		Name:         "Fleshpulper Giant",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(effects.WhenThisEnters("Fleshpulper Giant — you may destroy target creature with toughness 2 or less", effects.DestroyFirstLegalCardTarget),
				effects.TargetCreature("target creature with toughness 2 or less", effects.ToughnessLE(2))),
				"Fleshpulper Giant — destroy the creature?"),
		},
	})
}
