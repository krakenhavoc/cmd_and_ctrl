package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aligned Hedron Network — Artifact {4} (Battle for Zendikar, #5547):
//
//	"When this artifact enters, exile all creatures with power 5 or greater
//	 until this artifact leaves the battlefield. (Those creatures return under
//	 their owners' control.)"
//
// Each creature is exiled on the one "until this leaves" record (CR 610.3),
// so they all come back together, under their owners' control.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5a399467-abb7-4f33-962d-8bef5aa9ebf3",
		Name:         "Aligned Hedron Network",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepExile, Partial: true}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Aligned Hedron Network — exile all creatures with power 5 or greater until this artifact leaves the battlefield", effects.S82OldB314AlignedHedronNetwork),
		},
	})
}
