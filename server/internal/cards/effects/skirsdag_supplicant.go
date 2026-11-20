package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Skirsdag Supplicant — Creature — Human Cleric {2}{B}, 2/3 (Eldritch
// Moon, #5541):
//
//	"{B}, {T}, Discard a card: Each player loses 2 life."
//
// The discard is a cost, picked from hand at announce (CR 602.2b); the
// loss is Crushing Disappointment's eachPlayerLosesTwoLife, the
// activator included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c7d97311-6e34-4679-a678-e107022c785c",
		Name:         "Skirsdag Supplicant",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{B}, {T}, Discard a card: Each player loses 2 life.",
			Cost:  Plus(ManaCost("{B}"), TapCost(), DiscardACard()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return eachPlayerLosesTwoLife(NewContext(g, item))
			},
		}},
	})
}
