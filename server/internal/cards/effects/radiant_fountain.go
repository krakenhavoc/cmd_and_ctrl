package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "6db442e5-fbcc-4456-a4c5-bea1aee3fc8e",
		Name:         "Radiant Fountain",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Radiant Fountain — you gain 2 life", Do(GainLife{Amount: 2})),
		},
	})
}
