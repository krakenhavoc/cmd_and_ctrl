package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Misguided Rage — Sorcery {2}{R} (Scourge):
//
//	"Target player sacrifices a permanent of their choice."
//
// An edict over every permanent type, lands included. The player chooses
// as the spell resolves, and the sacrifice does not target (CR 701.21a), so
// hexproof and protection on their permanents do not matter.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f5c77355-96dc-46af-a7de-f547a1368ff1",
		Name:         "Misguided Rage",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			player, ok := effects.FirstLegalPlayerTarget(ctx)
			if !ok {
				return nil
			}
			ctx.Game.PlayerSacrificesForEffect(ctx.Source(), player, effects.SacrificeSpec("a permanent", effects.Permanent()),
				"Misguided Rage — sacrifice a permanent")
			return nil
		},
	})
}
