package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mind Warp — Sorcery {X}{3}{B} (Ice Age, S82 OLD, #5636):
//
//	"Look at target player's hand and choose X cards from it. That player discards those cards."
//
// Abandon Hope's look-and-choose with X as the count: only the caster sees the hand, and a hand with fewer
// than X cards is discarded whole (CR 608.2). With X = 0 the hand is only looked at.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5109aa76-ca40-44d1-abb4-5e7a23a2c274",
		Name:         "Mind Warp",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(game.TargetPurpose{Slot: 0, Discards: 1}),
		Targets:      effects.TargetPlayer("target player"),
		OnResolve:    effects.S82OldB844MindWarp,
	})
}
