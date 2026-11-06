package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dockside Extortionist — Creature — Goblin Pirate {1}{R}, 1/2 (Avatar: The
// Last Airbender Eternal, #5075):
//
//	"When this creature enters, create X Treasure tokens, where X is the
//	 number of artifacts and enchantments your opponents control."
//
// X is counted as the trigger resolves (CR 608.2h): every permanent an
// opponent of the trigger's controller controls that is an artifact, an
// enchantment or both, an artifact enchantment counted once (the ruling).
// The Treasures are created at once, the catalog's Treasure template.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "697bcfe1-ecbf-42a1-bfc7-0766d48ca56b",
		Name:         "Dockside Extortionist",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Dockside Extortionist — create a Treasure for each artifact and enchantment your opponents control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					n := len(MatchingBattlefield(ctx, And(Or(Artifact(), Enchantment()), OpponentControls())))
					if n == 0 {
						return nil
					}
					return CreateToken{Template: TreasureToken(), N: n}.Apply(ctx)
				}),
		},
	})
}
