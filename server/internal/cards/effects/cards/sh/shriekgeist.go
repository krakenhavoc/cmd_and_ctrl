package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shriekgeist — Creature — Spirit {1}{U}, 1/1 (Dark Ascension, #5571):
//
//	"Flying
//	 Whenever this creature deals combat damage to a player, that
//	 player mills two cards."
//
// The damaged player is the triggering event's target, as on Sword of Body
// and Mind.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e8dd4341-e7e5-49f7-b173-e4c3195e3334",
		Name:            "Shriekgeist",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Shriekgeist — that player mills two cards",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					return effects.MillCards{Player: ctx.Item.Trigger.Event.Target, N: 2}.Apply(ctx)
				}))),
		},
	})
}
