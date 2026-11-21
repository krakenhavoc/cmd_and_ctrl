package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Smashing Spree — Instant {1}{R} (Marvel Super Heroes Commander,
// #4455): "Target attacking creature gets +3/+3 and gains trample until
// end of turn." Built by trekB49AttackerPumpThreeTrample, shared with
// Run Amok, which prints the same text.
//
// No simplification.
func init() {
	effects.Register(effects.TrekB49AttackerPumpThreeTrample("f0371d3d-ca11-432a-bebd-684caad5781f", "Smashing Spree"))
}
