package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elementalist Adept — Creature — Human Wizard {1}{U}, 2/1
// (Foundations, #4465):
//
//	"Flash
//	 Prowess"
//
// Two printed keywords the engine enforces from PrintedKeywords: flash
// lets it be cast at instant speed, and prowess gives it +1/+1 until
// end of turn whenever its controller casts a noncreature spell.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "cf402270-ebc6-481b-b478-0e3acab7002a",
		Name:            "Elementalist Adept",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "prowess"},
	})
}
