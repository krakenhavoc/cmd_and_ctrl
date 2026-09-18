package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Exalted Flamer of Tzeentch — "At the beginning of your upkeep, return an
// instant or sorcery card at random from your graveyard to your hand.
// Whenever you cast an instant or sorcery spell, this creature deals
// 1 damage to each opponent."
func init() {
	Register(Spec{
		OracleID:     "6591a589-deb4-462f-a4f1-f6261f3433c3",
		Name:         "Exalted Flamer of Tzeentch",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, YouCast(Or(Instant(), Sorcery())), "Exalted Flamer of Tzeentch — 1 damage to each opponent", func(g *game.Game, item *game.StackItem) error {
				return damageToEachOpponent(g, item, 1)
			}),
			AtYourUpkeep("Exalted Flamer of Tzeentch — return a random instant or sorcery", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				ids := graveyardIDs(g, item.Controller, func(c game.Card) bool {
					return c.IsInstant() || c.IsSorcery()
				})
				pick := randomPick(ctx, ids, 1)
				if len(pick) == 0 {
					return nil
				}
				return g.ReturnFromGraveyardForEffect(pick[0], game.ZoneHand)
			}),
		},
	})
}
