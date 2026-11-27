package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Street Sweeper — Artifact Creature — Construct {6}, 4/6 (Return to Ravnica,
// #5566):
//
//	"Whenever this creature attacks, destroy all Auras attached to target land."
//
// The land is chosen as the trigger goes on the stack. On resolution every Aura
// attached to it right then is destroyed together (CR 701.8a), whoever controls
// the Aura; a land that has left makes the trigger do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "613d1a33-9fe5-48db-8277-955a055a787e",
		Name:         "Street Sweeper",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverThisAttacks("Street Sweeper — destroy all Auras attached to target land", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				land, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.DestroyAllMatching{Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
					return c.IsAura() && c.IsAttachedTo(land)
				}}.Apply(ctx)
			}), effects.TargetPermanent("target land", effects.Land())),
		},
	})
}
