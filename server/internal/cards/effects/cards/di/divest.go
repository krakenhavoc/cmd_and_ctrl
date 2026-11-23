package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Divest — Sorcery {B}:
//
//	"Target player reveals their hand. You choose an artifact or
//	 creature card from it. That player discards that card."
//
// The revealed-hand pick (ADR 0116): the whole table sees the hand
// (CR 701.20a), and only an artifact or creature card may be chosen (CR 701.9b).
// A hand with no artifact or creature card is revealed and nothing is discarded
// (CR 609.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9e0d7408-3827-4a65-9905-9fef3ae23c7e",
		Name:         "Divest",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve:    effects.TargetRevealsYouChooseDiscard(effects.Or(effects.Artifact(), effects.Creature()), "artifact or creature card"),
	})
}
