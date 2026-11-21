package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flare of Duplication — Instant {1}{R}{R} (Modern Horizons 3, #5948):
//
//	"You may sacrifice a nontoken red creature rather than pay this
//	 spell's mana cost.
//	 Copy target instant or sorcery spell. You may choose new targets for
//	 the copy."
//
// Reverberate's copy (CR 707.10) behind Flare of Fortitude's
// sacrifice-instead alternative cost, paid as the spell goes on the stack
// (CR 118.9).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f586ecb8-310c-4f40-a766-6a7ca46a254d",
		Name:         "Flare of Duplication",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetInstantOrSorcerySpell("target instant or sorcery spell"),
		AlternativeCosts: []game.AlternativeCost{
			effects.SacrificeInstead(1, "a nontoken red creature", 0, effects.Creature(), effects.OfColor("R"), effects.Not(effects.IsTokenPredicate())),
		},
		OnResolve: effects.CmdrB21CopyTheTargetSpell,
	})
}
