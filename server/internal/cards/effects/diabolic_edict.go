package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "058917c1-21ab-488a-9f9c-591c55f3c596",
		Name:         name,
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return targetPlayerSacrificesACreature(name+" — sacrifice a creature")(ctx.Game, item)
		},
	})
}
