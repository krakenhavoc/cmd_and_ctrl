package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shivan Gorge — Legendary Land (EDHREC rank 2363):
//
//	"{T}: Add {C}.
//	 {2}{R}, {T}: Shivan Gorge deals 1 damage to each opponent."
//
// A colourless land with a pinger. The mana half is an ordinary tap
// ability; the ping is a CR 602 activated ability sharing the tap,
// so a Gorge tapped for mana cannot ping the same turn. The Gorge
// itself is the damage source, so a damage doubler sees it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e90a1381-c9c1-4f57-928c-5d19dc065274",
		Name:         "Shivan Gorge",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label: "{2}{R}, {T}: Shivan Gorge deals 1 damage to each opponent.",
			Cost:  Plus(ManaCost("{2}{R}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return damageToEachOpponent(g, item, 1)
			},
		}},
	})
}
