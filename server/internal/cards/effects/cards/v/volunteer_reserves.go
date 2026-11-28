package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Volunteer Reserves — {1}{W} Creature — Human Soldier, 2/4 (Weatherlight, S82 OLD
// batch 772, #5629):
//
//	"Banding (Any creatures with banding, and up to one without, can attack in a
//	 band. Bands are blocked as a group. If any creatures with banding you control
//	 are blocking or being blocked by a creature, you divide that creature's combat
//	 damage, not its controller, among any of the creatures it's being blocked by
//	 or is blocking.)
//	 Cumulative upkeep {1}"
//
// Banding is a printed keyword the combat engine reads (CR 702.22, Benalish Hero's
// shape); the cumulative upkeep charges {1} for each age counter (CR 702.24a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a2dabd52-0b75-42d7-a07d-0df7b38646e2",
		Name:            "Volunteer Reserves",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordBanding},
		Triggered: []game.TriggeredAbility{
			effects.CumulativeUpkeep("Volunteer Reserves — cumulative upkeep {1}", "{1}"),
		},
	})
}
