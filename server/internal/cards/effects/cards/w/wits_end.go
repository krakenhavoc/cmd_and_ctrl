package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wit's End — Sorcery {5}{B}{B} (Dissension, #5598):
//
//	"Target player discards their hand."
//
// The whole hand goes at once (discardWholeHand); there is nothing to
// choose.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "97f11aed-7a5c-4a14-8c2c-5c230c1f05b4",
		Name:         "Wit's End",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetPlayer {
					continue
				}
				_, err := effects.DiscardWholeHand(ctx.Game, t.ID)
				return err
			}
			return nil
		},
	})
}
