package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Whitemane Lion — Creature — Cat {1}{W}, 2/2 (EDHREC rank 3786):
//
//	"Flash
//	 When this creature enters, return a creature you control to its
//	 owner's hand."
//
// The two-mana ETB re-buy: flash it in, bounce the creature whose
// enters trigger you want again — or bounce the Lion itself to save
// it from a wipe, which is a real and printed line. Flash rides
// PrintedKeywords; the bounce is a mandatory ETB trigger.
//
// "Return a creature you control" is a CHOICE, not a target — there
// is no "target" in the printed text — made on resolution
// (ReturnOneYouControl, the own_permanents prompt, #1214). It used to
// be a target clause picked when the trigger went on the stack, a
// declared simplification that let opponents see the choice before
// it resolved and excluded a creature with shroud; the
// resolution-time picker retired both. The Lion itself is always a
// candidate while it is still on the battlefield, and bouncing itself
// to save it from a wipe is a normal, sometimes correct, line.
func init() {
	Register(Spec{
		OracleID:        "e8d6084b-9b72-438e-a30a-851b888f3e4d",
		Name:            "Whitemane Lion",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Whitemane Lion — return a creature you control", Do(ReturnOneYouControl{
				Match:    MatchCreature,
				Question: "Whitemane Lion — return a creature you control to its owner's hand",
			})),
		},
	})
}
