package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bushmeat Poacher — Creature — Human Soldier {3}{B}, 2/4:
//
//	"{1}, {T}, Sacrifice another creature: You gain life equal to the
//	 sacrificed creature's toughness. Draw a card."
//
// The sacrifice is a cost, so the creature is gone before the ability
// resolves; its toughness is read from last-known information
// (departedCreatureToughness), the same read Consuming Vapors makes.
func init() {
	Register(Spec{
		OracleID:     "0287d541-3f73-4c8b-9a77-87f99898758d",
		Name:         "Bushmeat Poacher",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}, Sacrifice another creature: You gain life equal to the sacrificed creature's toughness. Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    Plus(ManaCost("{1}"), TapCost(), SacrificeAnotherN(1, "another creature", Creature())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if fed, ok := b17PermanentSacrificedToPay(g, item); ok {
					if t := departedCreatureToughness(g, fed); t > 0 {
						if err := (GainLife{Player: item.Controller, Amount: t}).Apply(ctx); err != nil {
							return err
						}
					}
				}
				return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
			},
		}},
	})
}
