package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Earthen Goo — Creature — Ooze {2}{R}, 2/2 (Coldsnap, S82 OLD b559):
//
//	"Trample
//	 Cumulative upkeep {R} or {G}
//	 This creature gets +1/+1 for each age counter on it."
//
// A hybrid cumulative upkeep (one {R/G} symbol per age counter) and a layer 7c
// pump that reads the Goo's age counters on every recompute.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f52c0ef9-ca35-47bd-9359-01e7b97ee770",
		Name:            "Earthen Goo",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static:          []game.StaticAbility{effects.OldB559PumpThisPerAgeCounter(1, 1)},
		Triggered: []game.TriggeredAbility{
			effects.CumulativeUpkeep("Earthen Goo — cumulative upkeep {R} or {G}", "{R/G}"),
		},
	})
}
