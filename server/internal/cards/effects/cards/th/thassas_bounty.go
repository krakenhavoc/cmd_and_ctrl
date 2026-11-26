package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thassa's Bounty — Sorcery {5}{U} (Theros):
//
//	"Draw three cards. Target player mills three cards."
//
// The instructions run in order (CR 608.2c): the controller draws, then the
// targeted player (any player, the controller included) mills. A target gone
// by resolution makes the spell do nothing (CR 608.2b), the draw included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c43425c1-ad01-4088-a061-f691fa2f6c20",
		Name:         "Thassa's Bounty",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 3},
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: effects.ResolveSteps(effects.DrawCards{N: 3}, effects.Instruction(func(ctx *effects.Context) error {
			return effects.TargetPlayerMillsEffect(3)(ctx.Game, ctx.Item)
		})),
	})
}
