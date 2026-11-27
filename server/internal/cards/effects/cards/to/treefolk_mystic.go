package cardsto

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Treefolk Mystic — Creature — Treefolk {3}{G} 2/4:
//
//	"Whenever this creature blocks or becomes blocked by a creature, destroy
//	 all Auras attached to that creature."
//
// One trigger for both halves (the block event names the pair). Every Aura
// attached to that creature is destroyed at once, whoever controls it.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "68a6fd00-9cb0-4518-97aa-b7f475da58c1",
		Name:         "Treefolk Mystic",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.SkewerSlingerFoe(ev, source.InstanceID) != uuid.Nil
			}, "Treefolk Mystic — destroy all Auras attached to that creature",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					foe := effects.SkewerSlingerFoe(item.Trigger.Event, item.SourceCardID)
					if foe == uuid.Nil {
						return nil
					}
					return effects.DestroyAllMatching{Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
						return c.IsAura() && c.AttachedTo.Kind == game.TargetCard && c.AttachedTo.ID == foe
					}}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
