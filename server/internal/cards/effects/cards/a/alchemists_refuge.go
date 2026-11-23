package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Alchemist's Refuge — Land:
//
//	"{T}: Add {C}.
//	 {G}{U}, {T}: You may cast spells this turn as though they had flash."
//
// Emergence Zone's permission without the sacrifice: GrantCastTiming with
// no window is "this turn", written onto the player and swept at cleanup
// (CR 514.2), so it outlives the land leaving the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "357ed28b-899f-404b-94ff-6fb2ef81d87b",
		Name:         "Alchemist's Refuge",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}{U}, {T}: You may cast spells this turn as though they had flash.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{G}{U}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.GrantCastTiming{
					Label: "You may cast spells this turn as though they had flash.",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
