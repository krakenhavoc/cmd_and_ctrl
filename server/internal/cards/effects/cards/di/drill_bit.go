package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Drill Bit — Sorcery {2}{B}:
//
//	"Spectacle {B} (You may cast this spell for its spectacle cost
//	 rather than its mana cost if an opponent lost life this turn.)
//	 Target player reveals their hand. You choose a nonland card from
//	 it. That player discards that card."
//
// Spectacle (CR 702.137a) is an alternative cost offered only while an
// opponent has lost life this turn: life lost, not a lower total, so
// damage counts and life gained back afterwards does not undo it (the
// rulings). The cost does not change what the spell does, and the
// mana value stays 3 either way. The pick is Thoughtseize's (ADR 0116).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "0dbd6e47-a8b4-4268-ba44-8924cd4963a6",
		Name:             "Drill Bit",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Spectacle("{B}")},
		Targets:          effects.TargetPlayer("target player"),
		OnResolve:        effects.TargetRevealsYouChooseDiscard(effects.Nonland(), "nonland card"),
	})
}
