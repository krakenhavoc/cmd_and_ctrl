package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dark Suspicions — Enchantment {2}{B}{B}:
//
//	"At the beginning of each opponent's upkeep, that player loses X
//	 life, where X is the number of cards in that player's hand minus
//	 the number of cards in your hand."
//
// Life LOSS, not damage: no prevention, no damage triggers. Both hands
// are counted as the trigger resolves (CR 608.2h); a negative X is zero
// (CR 107.1b), so an opponent holding no more cards than you loses
// nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cf5cf174-d976-4cd8-9721-c30fe784fb70",
		Name:         "Dark Suspicions",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtEachOpponentsUpkeep("Dark Suspicions — that player loses life equal to their hand size minus yours",
				func(g *game.Game, item *game.StackItem) error {
					player := effects.NewContext(g, item).Trigger().Event.Actor
					if player == uuid.Nil {
						return nil
					}
					x := effects.CardsInHand(g, player) - effects.CardsInHand(g, item.Controller)
					if x <= 0 {
						return nil
					}
					return g.ChangePlayerLifeForEffect(item.SourceCardID, player, -x)
				}),
		},
	})
}
