package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Guiding Bolt — Instant {2}{W} (Commander Legends: Battle for Baldur's Gate,
// S82 GAP #5948):
//
//	"Destroy target creature with power 4 or greater.
//	 Scry 2."
//
// Your Fate Ends Here's destroy-then-look: the target is destroyed if it is
// still legal as the spell resolves, and the caster scries whether or not
// anything was destroyed (an indestructible creature survives and the scry
// still happens). With the target gone the spell is countered by the rules
// (CR 608.2b) and does neither.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b433cf1f-a7f3-4b68-8bfb-e4196f6e3a12",
		Name:         "Guiding Bolt",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 2, Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature with power 4 or greater", effects.PowerGE(4)),
		OnResolve:    effects.DestroyTargetThenScry(2),
	})
}
