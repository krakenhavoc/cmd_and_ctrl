package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Overwhelming Denial — Instant {2}{U}{U} (Oath of the Gatewatch, #5545):
//
//	"Surge {U}{U} (You may cast this spell for its surge cost if you or a
//	 teammate has cast another spell this turn.)
//	 This spell can't be countered.
//	 Counter target spell."
//
// Surge is the shared gated alternative cost (CR 702.117a). "Can't be
// countered" is the Spec's CantBeCountered (Balustrade Wurm's shape); the
// counter is Counterspell's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "45aa8bdb-54da-4069-bd6c-c63396154464",
		Name:             "Overwhelming Denial",
		Completeness:     effects.CompletenessFull,
		CantBeCountered:  true,
		AlternativeCosts: []game.AlternativeCost{effects.Surge("{U}{U}")},
		Purpose:          game.Purpose{Answers: game.AnswerRemove},
		Targets:          effects.TargetSpell("target spell"),
		OnResolve:        effects.CounterTheTargetSpell,
	})
}
