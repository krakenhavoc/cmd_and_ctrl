package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Trapjaw Kelpie — Creature — Beast {4}{G/U}{G/U}, 3/3 (Eventide, #5590):
//
//	"Flash
//	 Persist (When this creature dies, if it had no -1/-1 counters on it,
//	 return it to the battlefield under its owner's control with a -1/-1
//	 counter on it.)"
//
// Flash (CR 702.8) and persist (CR 702.79).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0d6fccd5-ec14-4da6-972e-b2f5502818a4",
		Name:            "Trapjaw Kelpie",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", game.KeywordPersist},
	})
}
