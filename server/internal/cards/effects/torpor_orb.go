package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Torpor Orb — {2} Artifact:
//
//	"Creatures entering don't cause abilities to trigger."
//
// One trigger suppressor (#1735). It stops a creature's own "when this
// creature enters" and every "whenever a creature enters", whoever
// controls them, including evoke's "sacrifice it" (CR 702.74a), so an
// evoked creature stays. A noncreature permanent entering is
// untouched. So are replacement effects ("enters tapped", "enters with
// counters") and "as this enters" choices, because none of them is a
// triggered ability.
func init() {
	Register(Spec{
		OracleID:           "97326cad-b13c-4e52-82ce-850a39e5ff08",
		Name:               "Torpor Orb",
		Completeness:       CompletenessFull,
		TriggerSuppressors: []game.TriggerSuppressor{CreaturesEnteringDontTrigger("Torpor Orb")},
	})
}
