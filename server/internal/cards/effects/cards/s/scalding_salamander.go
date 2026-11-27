package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Scalding Salamander — Creature — Salamander {2}{R}, 2/1 (Exodus, S82 OLD
// b738, #5626):
//
//	"Whenever this creature attacks, you may have it deal 1 damage to each
//	 creature without flying defending player controls."
//
// Ronin Cliffrider's optional attack trigger with the flying filter: the set is
// the creatures the defending player controls that have no flying as the ability
// resolves (CR 608.2), read through the layers, and the Salamander is the source.
//
// No simplification.
func init() {
	trigger := effects.On(game.EventAttack, effects.ThisAttacked,
		"Scalding Salamander — you may have it deal 1 damage to each creature without flying defending player controls",
		func(g *game.Game, item *game.StackItem) error {
			if item.Trigger == nil {
				return nil
			}
			defender := effects.DefendingPlayerOfAttack(g, item.Trigger.Event)
			if defender == uuid.Nil {
				return nil
			}
			return effects.DamageEachMatching(effects.NewContext(g, item),
				effects.And(effects.Creature(), effects.WithoutKeyword("flying"), effects.ControlledBy(defender)), 1)
		})
	trigger.Purpose = game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, OpponentsOnly: true, Partial: true}}
	effects.Register(effects.Spec{
		OracleID:     "fcca3c9d-a2b6-4f29-b43c-2f061074a332",
		Name:         "Scalding Salamander",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(trigger, "Scalding Salamander — have it deal 1 damage to each creature without flying defending player controls?"),
		},
	})
}
