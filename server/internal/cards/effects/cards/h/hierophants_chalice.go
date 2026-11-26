package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hierophant's Chalice — Artifact {3} (Ixalan, S82 older sets):
//
//	"When this artifact enters, target opponent loses 1 life and you gain 1
//	 life.
//	 {T}: Add {C}."
//
// Skymarch Bloodletter's drain on a target opponent, and a colourless mana
// ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "52732f93-b951-43c3-bb02-c254996606e5",
		Name:         "Hierophant's Chalice",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "{T}: Add {C}",
		}},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Hierophant's Chalice — target opponent loses 1 life and you gain 1 life", effects.DrainTargetOpponentOne),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}
