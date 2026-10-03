package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Concordant Crossroads — World Enchantment {G} (EDHREC rank 1470):
//
//	"All creatures have haste."
//
// The one-mana haste enabler, for everyone. A Layer 6 keyword grant
// over every creature on the battlefield — every player's, as
// printed — so the engine's summoning-sickness checks (attacking,
// {T} abilities, mana abilities) read the granted haste exactly as
// they read a printed one.
//
// A world permanent: the world rule (CR 704.5k, ADR 0109 §8) puts it
// into its owner's graveyard when a newer one enters.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ff01b408-6d17-40a3-9efd-a1b341ec1307",
		Name:         "Concordant Crossroads",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, _ *game.Game, _ *game.Card) bool {
				return target.IsCreature()
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				for _, k := range c.Abilities {
					if k == "haste" {
						return
					}
				}
				c.Abilities = append(c.Abilities, "haste")
			},
		}},
	})
}
