package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Homarid Spawning Bed — Enchantment {U}{U} (Fallen Empires, S82 OLD batch 852,
// #5637):
//
//	"{1}{U}{U}, Sacrifice a blue creature: Create X 1/1 blue Camarid creature
//	 tokens, where X is the sacrificed creature's mana value."
//
// The sacrifice is paid as the ability is activated (a blue creature only), and X
// is the creature's mana value as it last existed on the battlefield
// (Morbid Curiosity's SacrificedManaValue, CR 608.2h): a blue token or a
// face-down creature is worth 0 and makes none. The Camarid is Homarid Spawning
// Bed's own token (no card was ever printed for it).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "173f2b8e-8072-4fdd-a469-8a1fb0bb6e64",
		Name:         "Homarid Spawning Bed",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{U}{U}, Sacrifice a blue creature: Create X 1/1 blue Camarid creature tokens, where X is the sacrificed creature's mana value.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{1}{U}{U}"), effects.SacrificeN(1, "a blue creature", effects.Creature(), effects.OfColor("U"))),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				n := effects.SacrificedManaValue(ctx)
				if n <= 0 {
					return nil
				}
				return effects.CreateToken{Template: effects.TokenCard("1/1 blue Camarid"), N: n}.Apply(ctx)
			},
		}},
	})
}
