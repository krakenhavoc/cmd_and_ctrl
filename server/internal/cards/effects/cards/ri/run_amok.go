package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Run Amok — Instant {1}{R} (Star Trek Commander, #4974):
//
//	"Target attacking creature gets +3/+3 and gains trample until end
//	 of turn."
//
// Smashing Spree's text; both are built by
// trekB49AttackerPumpThreeTrample.
//
// No simplification.
func init() {
	effects.Register(effects.TrekB49AttackerPumpThreeTrample("d463de75-8d3c-4077-a2cc-c304a63acc69", "Run Amok"))
}
