package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dispense Justice — Instant {2}{W} (Scars of Mirrodin, S82 OLD batch 469,
// #5577):
//
//	"Target player sacrifices an attacking creature of their choice.
//	 Metalcraft — That player sacrifices two attacking creatures of their
//	 choice instead if you control three or more artifacts."
//
// Wing Shards' edict, asked once or twice. Metalcraft is read as the spell
// resolves, and "two" is one run of two sacrifices (PlayerSacrificesNForEffect),
// so a player with a single attacker sacrifices just that one. The player
// chooses; nothing targets the creatures.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a47bf064-b810-43d0-89a5-13c107a317b0",
		Name:         "Dispense Justice",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			player, ok := effects.FirstLegalPlayerTarget(ctx)
			if !ok {
				return nil
			}
			n := 1
			if effects.OldB469Metalcraft(ctx.Game, ctx.Controller()) {
				n = 2
			}
			ctx.Game.PlayerSacrificesNForEffect(ctx.Source(), player,
				effects.SacrificeSpec("an attacking creature", effects.And(effects.Creature(), effects.IsAttackingCreature())),
				"Dispense Justice — sacrifice an attacking creature", n)
			return nil
		},
	})
}
