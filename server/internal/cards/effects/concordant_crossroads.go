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
// Sandbox simplification, declared: the WORLD supertype's "world
// rule" (CR 704.5k — when two world permanents are on the
// battlefield, the older one goes to the graveyard) is not
// modelled; no other world permanent is in the catalog for it to
// meet. Weaker for nobody today; the Crossroads would survive a
// second world enchantment where the printed card would not, and
// the caveat says so.
func init() {
	Register(Spec{
		OracleID:     "ff01b408-6d17-40a3-9efd-a1b341ec1307",
		Name:         "Concordant Crossroads",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The world rule isn't applied — it isn't put into the graveyard when another world enchantment enters."},
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
