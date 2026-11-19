package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:         "c43b616e-ef87-4c12-810b-a1a3260168fa",
		Name:             "Skewer the Critics",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Spectacle("{R}")},
		Targets:          TargetAny(),
		Purpose:          ForTargets(DamageToTarget(0, 3)),
		OnResolve:        damageToFirstTarget(3),
	})
}
