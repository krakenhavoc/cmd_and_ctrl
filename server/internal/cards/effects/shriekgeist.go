package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "e8dd4341-e7e5-49f7-b173-e4c3195e3334",
		Name:            "Shriekgeist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Shriekgeist — that player mills two cards",
				Do(Instruction(func(ctx *Context) error {
					return MillCards{Player: ctx.Item.Trigger.Event.Target, N: 2}.Apply(ctx)
				}))),
		},
	})
}
