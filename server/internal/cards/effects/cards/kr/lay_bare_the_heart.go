package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lay Bare the Heart — Sorcery {1}{B}:
//
//	"Target opponent reveals their hand. You choose a nonlegendary,
//	 nonland card from it. That player discards that card."
//
// The revealed-hand pick (ADR 0116): the whole table sees the hand
// (CR 701.20a), and only a nonlegendary, nonland card may be chosen (CR 701.9b).
// Nonlegendary is a card without the legendary supertype (CR 205.4a).
// A hand with no nonlegendary, nonland card is revealed and nothing is discarded
// (CR 609.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "910f4339-9125-45b7-929f-57e1adb6ac3c",
		Name:         "Lay Bare the Heart",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve:    effects.TargetRevealsYouChooseDiscard(effects.And(effects.Not(effects.Legendary()), effects.Nonland()), "nonlegendary, nonland card"),
	})
}
