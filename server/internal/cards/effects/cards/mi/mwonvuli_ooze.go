package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mwonvuli Ooze — {G} Creature — Ooze, 1+*/1+* (Weatherlight, S82 OLD batch 772,
// #5629):
//
//	"Cumulative upkeep {2}
//	 Mwonvuli Ooze's power and toughness are each equal to 1 plus twice the number
//	 of age counters on it."
//
// The cumulative upkeep charges {2} for each age counter (CR 702.24a). The power
// and toughness are a characteristic-defining ability (CR 604.3, layer 7a) read
// as the layers are recomputed: 1/1 with no counters, 3/3 with one, and so on.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "42b2c6d0-e0da-41b0-8876-f17dbe5db79a",
		Name:         "Mwonvuli Ooze",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.CumulativeUpkeep("Mwonvuli Ooze — cumulative upkeep {2}", "{2}"),
		},
		Static: []game.StaticAbility{
			effects.S82OldB772AgeCountersPT("Mwonvuli Ooze's power and toughness are each equal to 1 plus twice the number of age counters on it."),
		},
	})
}
