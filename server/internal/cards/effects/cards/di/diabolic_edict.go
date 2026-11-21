package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Diabolic Edict — Instant {1}{B} (Tempest, #5628):
//
//	"Target player sacrifices a creature of their choice."
//
// Chainer's Edict at instant speed and without flashback: the shared
// targetPlayerSacrificesACreature body (CR 701.21a). The target player
// chooses which creature as it resolves, so a creature with hexproof or
// protection is still a fine choice for them, and a player who controls
// no creature sacrifices nothing. A target player who left the game in
// response means the spell doesn't resolve (CR 608.2b).
//
// No simplification.
func init() {
	const name = "Diabolic Edict"
	effects.Register(effects.Spec{
		OracleID:     "058917c1-21ab-488a-9f9c-591c55f3c596",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.TargetPlayerSacrificesACreature(name+" — sacrifice a creature")(ctx.Game, item)
		},
	})
}
