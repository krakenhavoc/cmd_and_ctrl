package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Soulless One — Creature — Zombie Avatar {3}{B}, */*:
//
//	"Soulless One's power and toughness are each equal to the number of
//	 Zombies on the battlefield plus the number of Zombie cards in all
//	 graveyards."
//
// A layer 7a characteristic-defining ability, read live. Every Zombie
// counts, whoever controls it, Soulless One itself included; a card in a
// graveyard counts when it has the Zombie subtype.
//
// No simplification.
func init() {
	isZombie := func(c game.Card) bool { return c.HasSubtypeWord("Zombie") }
	Register(Spec{
		OracleID:     "27d98b94-a8a1-442b-8940-4ffab51b5164",
		Name:         "Soulless One",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Soulless One's power and toughness are each equal to the number of Zombies on the battlefield plus the number of Zombie cards in all graveyards.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, _ *game.Card) {
				n := cmdrB5CountBattlefield(g, func(x game.Card) bool { return x.IsCreature() && isZombie(x) }) +
					cmdrB5CountAllGraveyards(g, isZombie)
				c.Power, c.Toughness = n, n
			},
		}},
	})
}
