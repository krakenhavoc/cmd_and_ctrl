package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nivix, Aerie of the Firemind — Land:
//
//	"{T}: Add {C}.
//	 {2}{U}{R}, {T}: Exile the top card of your library. Until your next
//	 turn, you may cast it if it's an instant or sorcery spell."
//
// The spell-type gate is the permission's Filter, applied to the NAMED
// card (#3192, ADR 0066's second 2026-10-13 amendment): the grant
// covers the exiled card only if it is an instant or a sorcery, so a
// creature or a land exiled this way is just exiled. It is a cast
// permission only (a land is never cast), and the window is the
// engine's "until your next turn" duration, which closes as the
// holder's next turn begins.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9c482f1d-08b4-4882-918c-448a556d3fbe",
		Name:         "Nivix, Aerie of the Firemind",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{2}{U}{R}, {T}: Exile the top card of your library. Until your next turn, you may cast it if it's an instant or sorcery spell.",
			Cost:    Plus(ManaCost("{2}{U}{R}"), TapCost()),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: func(g *game.Game, item *game.StackItem) error {
				window := g.UntilYourNextTurnDuration(item.Controller)
				return ExileTopWithPermission{
					From: item.Controller, GrantTo: item.Controller, N: 1,
					CastOnly: true,
					Window:   &window,
					Filter:   game.PermissionFilter{InstantOrSorceryOnly: true},
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
