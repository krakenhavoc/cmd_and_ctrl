package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Formation — Instant {1}{W} (Ice Age, S82 OLD batch 837, #5636):
//
//	"Target creature gains banding until end of turn. (Any creatures with
//	 banding, and up to one without, can attack in a band. …)
//	 Draw a card at the beginning of the next turn's upkeep."
//
// Helm of Chatzuk's grant: banding is read off the layered ability list
// everywhere the engine asks for it (game/banding.go, ADR 0224). The delayed
// draw follows (CR 603.7a, next_upkeep_draw.go). With its one target gone the
// spell does not resolve, so nothing is drawn (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ad5d27ea-cead-4c95-8561-283855ce952e",
		Name:         "Formation",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, game.KeywordBanding)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.EbTargetCreatureGainsUntilEOT(game.KeywordBanding, "Formation — banding until end of turn")(ctx.Game, item); err != nil {
				return err
			}
			return effects.DrawACardAtTheNextUpkeep("Formation — draw a card").Apply(ctx)
		},
	})
}
