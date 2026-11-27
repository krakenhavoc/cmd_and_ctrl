package cardsto

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Twisted Justice — Sorcery {4}{U}{B} (Ravnica: City of Guilds):
//
//	"Target player sacrifices a creature of their choice. You draw cards equal to
//	 that creature's power."
//
// Consume's edict: the target player picks the creature as the spell resolves,
// and the draw waits for the sacrifice (a keyed continuation). The power is the
// creature's as it last existed on the battlefield (CR 608.2h), and a creature
// with no power, or none to sacrifice, draws nothing. The caster, not the
// target, draws.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "89c5dec1-cefe-4ce3-a9df-b5cb9fa73b35",
		Name:         "Twisted Justice",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Draws: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			victim, ok := effects.FirstLegalPlayerTarget(ctx)
			if !ok {
				return nil
			}
			return ctx.Game.PlayerSacrificesThenKeyedForEffect(ctx.Source(), victim,
				effects.SacrificeSpec("a creature", effects.Creature()),
				"Twisted Justice — sacrifice a creature", 1,
				effects.OldB579DrawEqualToSacrificedPower, []uuid.UUID{ctx.Controller(), victim})
		},
	})
}
