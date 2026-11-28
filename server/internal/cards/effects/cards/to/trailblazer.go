package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Trailblazer — Instant {2}{G}{G} (Ice Age, S82 OLD, #5636):
//
//	"Target creature can't be blocked this turn."
//
// Rogue's Passage's restriction as a spell: the target can't be blocked until end of turn, even if it
// isn't attacking yet.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "53ae0359-aa3c-464b-bdbd-139ae17ebd76",
		Name:         "Trailblazer",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.CantBeBlockedUntilEOT(0, 0, 0)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.S82OldB844Trailblazer,
	})
}
