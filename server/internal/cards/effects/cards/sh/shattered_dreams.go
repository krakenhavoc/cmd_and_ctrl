package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shattered Dreams — Sorcery {B}:
//
//	"Target opponent reveals their hand. You choose an artifact card
//	 from it. That player discards that card."
//
// The revealed-hand pick (ADR 0116): the whole table sees the hand
// (CR 701.20a), and only an artifact card may be chosen (CR 701.9b).
// A hand with no artifact card is revealed and nothing is discarded
// (CR 609.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "91c2f488-2b49-4499-8d5d-c1aff73a4b57",
		Name:         "Shattered Dreams",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve:    effects.TargetRevealsYouChooseDiscard(effects.Artifact(), "artifact card"),
	})
}
