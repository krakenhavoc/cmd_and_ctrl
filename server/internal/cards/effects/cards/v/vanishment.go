package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vanishment — Instant {4}{U} (Avacyn Restored, S82 OLD, #5570):
//
//	"Put target nonland permanent on top of its owner's library.
//	 Miracle {U} (You may cast this card for its miracle cost when you draw
//	 it if it's the first card you drew this turn.)"
//
// Time Ebb's body on any nonland permanent, with Thunderous Wrath's miracle
// declaration; a token put there ceases to exist.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "a2ae1a1b-5c4f-46ed-90da-bd50ca914e7c",
		Name:             "Vanishment",
		Completeness:     effects.CompletenessFull,
		Targets:          effects.TargetPermanent("target nonland permanent", effects.Nonland()),
		AlternativeCosts: []game.AlternativeCost{effects.Miracle("{U}")},
		OnResolve:        effects.PutTargetOnTopOfOwnersLibrary,
	})
}
