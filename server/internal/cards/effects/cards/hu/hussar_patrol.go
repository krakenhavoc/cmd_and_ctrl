package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hussar Patrol — Creature — Human Knight {2}{W}{U}, 2/4 (Return to Ravnica, #5566):
//
//	"Flash (You may cast this spell any time you could cast an
//	 instant.) / Vigilance"
//
// The printed keywords ride PrintedKeywords; the engine enforces them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "bd73581c-c4bd-46cb-8004-75aeafa24288",
		Name:            "Hussar Patrol",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "vigilance"},
	})
}
