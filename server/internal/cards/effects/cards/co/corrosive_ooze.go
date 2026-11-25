package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Corrosive Ooze — Creature — Ooze {1}{G}, 2/2 (Dominaria, S82 OLD batch 214,
// #5530):
//
//	"Whenever this creature blocks or becomes blocked by an equipped creature,
//	 destroy all Equipment attached to that creature at end of combat."
//
// Ornery Goblin's per-pair block trigger (CR 509.3b, 509.3d), applying only
// when the other creature of the pair is equipped as the block is declared. The
// ability schedules a CR 603.7 delayed trigger at the beginning of the end of
// combat step, naming that creature; the Equipment destroyed are the ones
// attached to it then (not those it had when the block was declared), and a
// creature that has left has none.
//
// No simplification.
func init() {
	const label = "Corrosive Ooze — destroy all Equipment attached to that creature at end of combat"
	effects.Register(effects.Spec{
		OracleID:     "416cca4a-b1b2-49d5-9cac-ec347056bdb0",
		Name:         "Corrosive Ooze",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				foe := effects.SkewerSlingerFoe(ev, source.InstanceID)
				if foe == uuid.Nil {
					return false
				}
				c, ok := g.LookupCardForEffect(foe)
				return ok && effects.IsEquippedCreature(g, c)
			}, label, func(g *game.Game, item *game.StackItem) error {
				if item.Trigger == nil {
					return nil
				}
				foe := effects.SkewerSlingerFoe(item.Trigger.Event, item.SourceCardID)
				if foe == uuid.Nil {
					return nil
				}
				return effects.ScheduleDelayedTrigger{
					At:    game.StepEndCombat,
					Label: label,
					Cards: []uuid.UUID{foe},
					Body:  effects.OldB214DestroyAttachedEquipmentBody,
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
