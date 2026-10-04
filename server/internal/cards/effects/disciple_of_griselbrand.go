package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Disciple of Griselbrand — Creature — Human Cleric {1}{B}, 1/1:
//
//	"{1}, Sacrifice a creature: You gain life equal to the sacrificed
//	 creature's toughness."
//
// The sacrificed creature's toughness is last-known information, read
// back after the cost has taken it (see Bushmeat Poacher).
func init() {
	Register(Spec{
		OracleID:     "2d92a035-dd7a-4426-a8c0-f04e0b836dad",
		Name:         "Disciple of Griselbrand",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{1}, Sacrifice a creature: You gain life equal to the sacrificed creature's toughness.",
			Cost:  Plus(ManaCost("{1}"), SacrificeACreature()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				fed, ok := b17PermanentSacrificedToPay(g, item)
				if !ok {
					return nil
				}
				return GainLife{Player: item.Controller, Amount: departedCreatureToughness(g, fed)}.Apply(NewContext(g, item))
			},
		}},
	})
}
