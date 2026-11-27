package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Screeching Silcaw — Creature — Bird {1}{U}, 1/2 (Scars of Mirrodin, S82 OLD
// batch 469, #5577):
//
//	"Flying
//	 Metalcraft — Whenever this creature deals combat damage to a player, if
//	 you control three or more artifacts, that player mills four cards."
//
// Shriekgeist's combat-damage trigger with the metalcraft "if" as part of its
// condition and checked again as it resolves (CR 603.4); the damaged player
// is the triggering event's target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7a1a4fe7-6574-4242-ae6f-108c1b9a5635",
		Name:            "Screeching Silcaw",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, effects.AllOf(effects.ThisDealtCombatDamageToAPlayer, effects.OldB469WhenMetalcraft),
				"Screeching Silcaw — that player mills four cards",
				effects.OldB469IfMetalcraft(effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					return effects.MillCards{Player: ctx.Item.Trigger.Event.Target, N: 4}.Apply(ctx)
				})))),
		},
	})
}
