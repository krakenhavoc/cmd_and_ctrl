package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// River Herald Guide — Creature — Merfolk Scout {2}{G}, 3/1 (The Lost
// Caverns of Ixalan, #4470):
//
//	"Vigilance
//	 When this creature enters, it explores."
//
// Vigilance is the printed keyword (CR 702.20a). The enters trigger is
// River Herald Scout's: the object that entered explores
// (thatCreatureExplores, CR 701.44a), by its last known information if it
// has left (CR 701.44c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b9687fef-4cb2-4bf1-91ad-49dada07f4d6",
		Name:            "River Herald Guide",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("River Herald Guide — it explores", thatCreatureExplores),
		},
	})
}
