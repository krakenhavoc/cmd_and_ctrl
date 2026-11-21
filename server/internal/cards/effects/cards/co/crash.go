package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crash — Instant {2}{R} (Mercadian Masques, #5619):
//
//	"You may sacrifice a Mountain rather than pay this spell's mana
//	 cost.
//	 Destroy target artifact."
//
// Shatter's target and body, with Fireblast's sacrifice alternative cost
// at one Mountain (SacrificeInstead, #1727). The Mountain is sacrificed
// as a cost with Crash already on the stack (CR 601.2h), so a countered
// Crash still costs it. Any land with the subtype Mountain pays it, a
// nonbasic one included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "562af5da-1080-46d2-8a38-71f06b7e8457",
		Name:         "Crash",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target artifact", effects.Artifact()),
		AlternativeCosts: []game.AlternativeCost{
			effects.SacrificeInstead(1, "a Mountain", 0, effects.HasSubtype("Mountain")),
		},
		OnResolve: effects.DestroyTheTargetPermanent,
	})
}
