package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skewer the Critics — Sorcery {2}{R} (Breaking News, #5929):
//
//	"Spectacle {R} (You may cast this spell for its spectacle cost rather
//	 than its mana cost if an opponent lost life this turn.)
//	 Skewer the Critics deals 3 damage to any target."
//
// Spectacle is the shared alternative cost (CR 702.137a, Spectacle); the
// body is Fiery Temper's three damage to any target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "c43b616e-ef87-4c12-810b-a1a3260168fa",
		Name:             "Skewer the Critics",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Spectacle("{R}")},
		Targets:          effects.TargetAny(),
		Purpose:          effects.ForTargets(effects.DamageToTarget(0, 3)),
		OnResolve:        effects.DamageToFirstTargetStep(3),
	})
}
