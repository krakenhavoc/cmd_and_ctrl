package cardsbu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Celestial Flare — Instant {W}{W} (Magic Origins, S82 older sets):
//
//	"Target player sacrifices an attacking or blocking creature of their
//	 choice."
//
// Crude Bent Blade's "target player sacrifices" over the creatures that are
// attacking or blocking as the spell resolves; the player chooses (CR 701.21a).
// A target gone by resolution makes the spell do nothing (CR 608.2b); with no
// attacking or blocking creature nothing is sacrificed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "efdaa3eb-e5d1-4438-b21d-2ec0fff44721",
		Name:         "Celestial Flare",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			p := effects.TargetedPlayerOf(ctx)
			if p == uuid.Nil {
				return nil
			}
			ctx.Game.PlayerSacrificesForEffect(item.SourceCardID, p,
				effects.SacrificeSpec("an attacking or blocking creature", effects.And(effects.Creature(), effects.AttackingOrBlocking())),
				"Celestial Flare — sacrifice an attacking or blocking creature")
			return nil
		},
	})
}
