package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Master of the Veil — 2/3 Human Wizard for {2}{U}{U}:
//
//	"Morph {2}{U}
//	 When this creature is turned face up, you may turn target
//	 creature with a morph ability face down."
//
// The card that closes the loop, and the reason #1209 and #1194 have
// to agree on one face-down model rather than two. Cast the Master
// face down for {3}, turn it up for {2}{U} (CR 708.6, #1194's special
// action), and its trigger (CR 708.8, watching EventTurnedFaceUp)
// hides another morph — which its controller can then turn face up
// again for ITS morph cost, because CR 702.37e keys the permission on
// the card having morph rather than on how the permanent got face
// down. It can even target the Master itself, which is the loop the
// card is built around.
//
// Both directions are the engine's: `Morph` is the CR 708.4
// alternative cost (#1194) and the trigger's body is #1209's
// primitive. The card file is the two lines that vary.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ee9104d5-3583-4549-8dd0-5b69274f4b4f",
		Name:         "Master of the Veil",
		Completeness: CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{
			Morph("{2}{U}"),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(
				Optional(
					WhenThisIsTurnedFaceUp(
						"Master of the Veil — turn a creature with morph face down",
						turnSingleTargetFaceDown),
					"Master of the Veil — turn a creature with a morph ability face down?"),
				TargetCreature("target creature with a morph ability", WithMorphAbility())),
		},
	})
}
