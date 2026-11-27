package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ancient Runes — Enchantment {2}{R} (Tempest, S82 OLD batch 757, #5628):
//
//	"At the beginning of each player's upkeep, this enchantment deals
//	 damage to that player equal to the number of artifacts they control."
//
// Molder Slug's every-upkeep trigger; the player is the upkeep's, and the
// artifacts are counted as the trigger resolves (CR 608.2h) through the
// layered view, so an animated artifact counts. With none, no damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1a128c02-febf-4ce6-b4d1-28274b5c4ff0",
		Name:         "Ancient Runes",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtEachUpkeep("Ancient Runes — damage to that player equal to the artifacts they control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					p := ctx.Trigger().Event.Actor
					n := effects.B03ArtifactsControlled(g, p)
					if p == uuid.Nil || n <= 0 {
						return nil
					}
					return effects.DealDamage{Source: item.SourceCardID, Target: p, Amount: n}.Apply(ctx)
				}),
		},
	})
}
