package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Adventurer's Inn — Land — Town (EDHREC rank 1551):
//
//	"When this land enters, you gain 2 life.
//	 {T}: Add {C}."
//
// Radiant Fountain with a Town subtype, which the printed type line
// carries for anything that cares. The life is a real ETB trigger
// with a response window, as printed; the land enters untapped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "232bd88c-ecdb-43dd-b34a-d381cb3bedf2",
		Name:         "Adventurer's Inn",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Adventurer's Inn — you gain 2 life", effects.Do(effects.GainLife{Amount: 2})),
		},
	})
}
