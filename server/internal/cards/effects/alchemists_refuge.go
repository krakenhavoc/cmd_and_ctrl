package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "357ed28b-899f-404b-94ff-6fb2ef81d87b",
		Name:         "Alchemist's Refuge",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{G}{U}, {T}: You may cast spells this turn as though they had flash.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{G}{U}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return GrantCastTiming{
					Label: "You may cast spells this turn as though they had flash.",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
