package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sabertooth Cobra — Creature — Snake {2}{G}, 2/2 (Mirage, #5633):
//
//	"Whenever this creature deals damage to a player, that player gets a poison
//	 counter. The player gets another poison counter at the beginning of their
//	 next upkeep unless they pay {2} before that step."
//
// The second counter is a delayed trigger at that player's next upkeep carrying
// the payment (#6220): they may pay {2} any time they have priority until that
// step begins, and paying removes it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "115c7ca7-c27b-43ca-9281-be6913e90059",
		Name:         "Sabertooth Cobra",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, effects.ThisDealtDamageToAPlayer, "Sabertooth Cobra — that player gets a poison counter, and another at their next upkeep unless they pay {2}",
				func(g *game.Game, item *game.StackItem) error {
					victim := item.Trigger.Event.Target
					if p := g.PlayerByIDForEffect(victim); p == nil || p.Eliminated {
						return nil
					}
					ctx := effects.NewContext(g, item)
					if err := g.AddPlayerCounterByForEffect(item.Controller, victim, game.CounterPoison, 1); err != nil {
						return err
					}
					return effects.ScheduleDelayedTrigger{
						Label:       "Sabertooth Cobra — another poison counter unless they pay {2}",
						At:          game.StepUpkeep,
						TurnOf:      victim,
						Body:        sabertoothCobraPoison,
						Params:      game.EffectParams{Player: victim},
						PayToCancel: &game.PayToCancel{Player: victim, Cost: "{2}"},
					}.Apply(ctx)
				}),
		},
	})
}

// sabertoothCobraPoison is the delayed trigger: the player gets another poison
// counter. An on-disk key: never renamed, never reused.
var sabertoothCobraPoison = game.DelayedBody("old-b793/sabertooth-cobra-poison",
	func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
		if p.Player == uuid.Nil {
			return nil
		}
		if pl := g.PlayerByIDForEffect(p.Player); pl == nil || pl.Eliminated {
			return nil
		}
		return g.AddPlayerCounterByForEffect(item.Controller, p.Player, game.CounterPoison, 1)
	})
