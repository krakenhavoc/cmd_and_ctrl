package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Radiant Fountain — Land (EDHREC rank 1573):
//
//	"When this land enters, you gain 2 life.
//	 {T}: Add {C}."
//
// The life is a real ETB trigger with a response window, as printed
// — the gain-land shape with the number doubled and no colour. The
// land enters untapped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6db442e5-fbcc-4456-a4c5-bea1aee3fc8e",
		Name:         "Radiant Fountain",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Radiant Fountain — you gain 2 life", effects.Do(effects.GainLife{Amount: 2})),
		},
	})
}
