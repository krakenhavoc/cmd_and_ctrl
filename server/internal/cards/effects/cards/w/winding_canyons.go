package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Winding Canyons — Land (Weatherlight, S82 older sets batch 780, #5629):
//
//	"{T}: Add {C}.
//	 {2}, {T}: You may cast creature spells this turn as though they had flash."
//
// Emergence Zone's timing grant narrowed to creature spells: a statement on the controller (game.GrantCastTiming) that
// lasts for the rest of the turn.
//
// No simplification.
func init() {
	const label = "You may cast creature spells this turn as though they had flash."
	effects.Register(effects.Spec{
		OracleID:     "622e2561-48b1-4aca-9abb-9a3c284dcceb",
		Name:         "Winding Canyons",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: " + label,
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.GrantCastTiming{
					Timing: game.TimingFlash,
					Filter: game.PermissionFilter{CreatureOnly: true},
					Label:  label,
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
