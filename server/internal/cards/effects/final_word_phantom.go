package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Final-Word Phantom — Creature — Spirit Detective {2}{U}, 1/4 (Murders
// at Karlov Manor Commander, #5368):
//
//	"Flash
//	 Flying
//	 During each opponent's end step, you may cast spells as though they
//	 had flash."
//
// A cast-timing declaration (Vedalken Orrery's) gated on the phase: it
// is open only in the end step of a turn an opponent is taking, read live
// on every query, so it never opens in your own end step. It is about
// casting, so it does not let its controller play a land at instant
// speed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f07db0f7-7d25-4395-9af3-6b0ca9197b61",
		Name:            "Final-Word Phantom",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		CastTimings: []game.CastTimingRule{
			cmdr2S22CastDuringOpponentsEndStep("During each opponent's end step, you may cast spells as though they had flash."),
		},
	})
}
