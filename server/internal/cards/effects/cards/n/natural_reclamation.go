package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Natural Reclamation — Instant {4}{G} (Commander Legends: Battle for
// Baldur's Gate, #5948):
//
//	"Cascade
//	 Destroy target artifact or enchantment."
//
// Cascade is the catalog keyword (it triggers on cast and resolves before the
// spell, CR 702.85a), as on Bituminous Blast; the rest is Revoke Existence's
// destroy over an artifact or an enchantment.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "98266582-12fd-4a7b-a9bb-e9cc4370ed28",
		Name:         "Natural Reclamation",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
		Triggered:    []game.TriggeredAbility{effects.Cascade()},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyFirstLegalCardTarget(ctx.Game, item)
		},
	})
}
