package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flow of Maggots — Creature — Insect {2}{B}, 2/2 (Ice Age, S82 OLD batch 838,
// #5636):
//
//	"Cumulative upkeep {1}
//	 This creature can't be blocked by non-Wall creatures."
//
// Juggernaut's block rule inverted: only a Wall may block it. A blocker's
// types are read as blockers are declared (CR 509.1b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2d6a387f-7ad4-419e-951a-9dcd4c9ac823",
		Name:         "Flow of Maggots",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.CumulativeUpkeep("Flow of Maggots — cumulative upkeep {1}", "{1}")},
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedExceptBy(effects.OnSelf(), effects.OfCreatureType("Wall"), "Walls"),
		},
	})
}
