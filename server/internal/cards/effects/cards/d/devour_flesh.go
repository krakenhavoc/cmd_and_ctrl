package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Devour Flesh — Instant {1}{B} (Gatecrash, S82 OLD b408, #5565):
//
//	"Target player sacrifices a creature of their choice, then gains life equal
//	 to that creature's toughness."
//
// Tribute to Hunger's edict (PlayerSacrificesThenKeyedForEffect) with Consuming
// Vapors' continuation, whose carry names who gains: here the targeted player
// is both the one who sacrifices and the one who gains, so the toughness is
// the sacrificed creature's as it last existed (CR 608.2h). A player with no
// creature sacrifices nothing and gains nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6c40f825-836e-4b76-8e43-267b4abf013b",
		Name:         "Devour Flesh",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			victim, ok := effects.FirstLegalPlayerTarget(ctx)
			if !ok {
				return nil
			}
			return ctx.Game.PlayerSacrificesThenKeyedForEffect(ctx.Source(), victim,
				effects.SacrificeSpec("a creature", effects.Creature()),
				"Devour Flesh — sacrifice a creature", 1,
				effects.ConsumingVaporsGain, []uuid.UUID{victim, victim})
		},
	})
}
