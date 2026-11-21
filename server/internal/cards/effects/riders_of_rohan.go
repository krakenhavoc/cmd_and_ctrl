package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Riders of Rohan — Creature — Human Knight {3}{R}{W}, 4/4 (Tales of
// Middle-earth Commander):
//
//	"When this creature enters, create two 2/2 red Human Knight creature
//	 tokens with trample and haste.
//	 Dash {4}{R}{W} (You may cast this spell for its dash cost. If you do,
//	 it gains haste, and it's returned from the battlefield to its owner's
//	 hand at the beginning of the next end step.)"
//
// Dash is the shared constructor (freerunning.go); the entry trigger
// fires on a dashed cast too, since the creature still enters.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "688a559d-249f-4d03-98e6-581843e7ddc3",
		Name:             "Riders of Rohan",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Dash("{4}{R}{W}")},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Riders of Rohan — create two 2/2 red Human Knight tokens with trample and haste",
				Do(CreateToken{Template: TokenCard("2/2 red Human Knight with trample and haste"), N: 2})),
		},
	})
}
