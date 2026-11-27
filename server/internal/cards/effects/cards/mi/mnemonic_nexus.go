package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mnemonic Nexus — Instant {3}{U} (Ravnica: City of Guilds, S82 OLD batch 581,
// #5600):
//
//	"Each player shuffles their graveyard into their library."
//
// Each seat, in APNAP order (CR 101.4), moves its graveyard into its library
// and shuffles it, even when the graveyard was empty (the shared
// FinaleShuffleGraveyardIntoLibrary body: Blessed Respite's).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6fe302f2-5594-478c-a66b-89b4b96e3611",
		Name:         "Mnemonic Nexus",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, seat := range effects.ApnapPlayers(ctx.Game) {
				if err := effects.FinaleShuffleGraveyardIntoLibrary(ctx, seat); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
