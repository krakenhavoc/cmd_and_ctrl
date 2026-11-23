package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reverberate — Instant {R}{R}:
//
//	"Copy target instant or sorcery spell. You may choose new
//	 targets for the copy."
//
// The plainest CR 707.10 card there is, and the reason the
// primitive exists. Note the target clause has no "you control":
// Reverberate's whole job in Commander is to answer the table's
// best spell by taking a second one for yourself — copy the
// opponent's Time Warp, copy the Expropriate, copy the tutor.
//
// The copy is controlled by Reverberate's controller, not by the
// controller of the spell it copied (CR 707.10). That asymmetry is
// carried by CopySpell.Controller defaulting to ctx.Controller().
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a1f55890-31c5-4ed4-a2cd-7a4a9f05f8ca",
		Name:         "Reverberate",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetInstantOrSorcerySpell("target instant or sorcery spell"),
		OnResolve:    effects.CmdrB21CopyTheTargetSpell,
	})
}
