package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dead-Iron Sledge — Artifact — Equipment {1} (Mirrodin, #5606):
//
//	"Whenever equipped creature blocks or becomes blocked by a creature,
//	 destroy both creatures.
//	 Equip {2}"
//
// Lim-Dûl's Cohort's per-pair block trigger (CR 509.3b, 509.3d) read
// off the Equipment: one block event per blocker–attacker pair, so an
// equipped creature blocked by two creatures triggers twice. "Both
// creatures" are the two creatures of the pair the trigger saw, the
// equipped one and the other, destroyed at the same time as the trigger
// resolves, wherever the Sledge is by then. Either may regenerate or be
// indestructible; one that has left the battlefield is not destroyed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e1ee79de-e4c3-48f9-87a7-18f72e2f6116",
		Name:         "Dead-Iron Sledge",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Target != uuid.Nil && (source.IsAttachedTo(ev.CardID) || source.IsAttachedTo(ev.Target))
			}, "Dead-Iron Sledge — destroy both creatures", func(g *game.Game, item *game.StackItem) error {
				ev := item.Trigger.Event
				g.DestroyPermanentsForEffect([]uuid.UUID{ev.CardID, ev.Target})
				return nil
			}),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
