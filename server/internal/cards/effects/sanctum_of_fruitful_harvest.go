package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sanctum of Fruitful Harvest — Legendary Enchantment — Shrine {2}{G}:
//
//	"At the beginning of your first main phase, add X mana of any one
//	 color, where X is the number of Shrines you control."
//
// A triggered ability that adds mana (not a mana ability — it has no
// cost and uses the stack), so the mana goes in through
// AddManaWithOptionsForEffect, the Dark Ritual path, as #742's one
// pick of X tokens. The printed text says "any one color" with no
// commander-identity clause, so the pick opts out of the identity
// narrowing and offers all five colours. X is counted when the trigger
// resolves, so a Shrine that left in response is not counted. The
// Sanctum is a Shrine itself. The mana empties from the pool at the
// end of the main phase like any other (CR 106.4).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "132859dd-de66-45c6-8af4-ab5e202a17b0",
		Name:         "Sanctum of Fruitful Harvest",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourPrecombatMain("Sanctum of Fruitful Harvest — add X mana of any one color", func(g *game.Game, item *game.StackItem) error {
				shrines := 0
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.Controller == item.Controller && c.HasSubtype("Shrine") {
						shrines++
					}
				}
				return g.AddManaWithOptionsForEffect(item.Controller, item.SourceCardID, OneColorOfAmount(shrines),
					game.AddManaOptions{IgnoreCommanderIdentity: true})
			}),
		},
	})
}
