package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Loyal Gyrfalcon — Creature — Bird {3}{W}, 3/3 (Eventide, #5590):
//
//	"Defender, flying
//	 Whenever you cast a white spell, this creature loses defender until end
//	 of turn."
//
// Jaded Analyst's layer 6 removal of defender until end of turn, on a
// white-spell cast trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f693cf76-0919-4b89-92de-e619be006abb",
		Name:            "Loyal Gyrfalcon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender", "flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.OfColor("W"), "Loyal Gyrfalcon — loses defender until end of turn",
				effects.ThisLosesKeywordUntilEndOfTurn("defender", "Loyal Gyrfalcon — loses defender until end of turn")),
		},
	})
}
