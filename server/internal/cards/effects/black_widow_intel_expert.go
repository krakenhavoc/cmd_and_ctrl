package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Black Widow, Intel Expert — Legendary Creature — Human Spy Hero {2}{W}, 3/3
// (Marvel Super Heroes Commander, #4455):
//
//	"First strike
//	 Whenever Black Widow deals combat damage to an opponent, you and that
//	 player each draw two cards."
//
// Hydra Omnivore's condition (b38ThisDealtCombatDamageToAnOpponent): this
// creature itself, on a player who is an opponent. "That player" is the
// player the damage event named, read off the carried trigger context, so a
// trigger that waits on the stack still draws for the right opponent. The
// controller draws first, then that player.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "99c3a315-db53-4292-9d31-1efd791d3a13",
		Name:            "Black Widow, Intel Expert",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b38ThisDealtCombatDamageToAnOpponent(ev, source, g)
			}, "Black Widow, Intel Expert — you and that player each draw two cards",
				Do(DrawCards{N: 2}, Instruction(func(ctx *Context) error {
					return DrawCards{Player: ctx.Item.Trigger.Event.Target, N: 2}.Apply(ctx)
				}))),
		},
	})
}
