package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hushbringer — {1}{W} Creature — Faerie 1/2:
//
//	"Flying, lifelink
//	 Creatures entering or dying don't cause abilities to trigger."
//
// Two trigger suppressors (#1735): Torpor Orb's line, plus
// SuppressesDying for the second half. A dies trigger looks back in
// time (CR 603.10a), so a Hushbringer that dies in the same wipe as
// the other creatures still stops their dies triggers, and it stops
// the triggers of its own death too. A creature that leaves the
// battlefield any other way (exiled, bounced) is not dying, and its
// leave triggers still happen.
func init() {
	dying := SuppressesDying(Creature())
	dying.Label = "Hushbringer"
	Register(Spec{
		OracleID:        "3d21f710-6bbc-41ae-a8e5-02debe3e02bd",
		Name:            "Hushbringer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink"},
		TriggerSuppressors: []game.TriggerSuppressor{
			CreaturesEnteringDontTrigger("Hushbringer"),
			dying,
		},
	})
}
