package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gleaming Geardrake — Artifact Creature — Drake {U}{R}, 1/1 (Murders at
// Karlov Manor, #4469):
//
//	"Flying
//	 When this creature enters, investigate. (Create a Clue token. It's an
//	 artifact with "{2}, Sacrifice this token: Draw a card.")
//	 Whenever you sacrifice an artifact, put a +1/+1 counter on this
//	 creature."
//
// The enters trigger is Novice Inspector's one Clue. The sacrifice
// trigger is Crime Novelist's condition (b12YouSacrificedAnArtifact),
// which reads the sacrifice event while the permanent is still on the
// battlefield, so a cracked Clue or Treasure counts. The counter goes on
// the Geardrake only while it is still there: sacrificing the Geardrake
// itself triggers it, and the counter then has nowhere to go.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "af294df7-f8a0-4d52-88cb-086678463970",
		Name:            "Gleaming Geardrake",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Gleaming Geardrake — investigate", Do(CreateToken{Template: ClueToken(), N: 1})),
			On(game.EventSacrifice, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b12YouSacrificedAnArtifact(ev, source, g)
			}, "Gleaming Geardrake — put a +1/+1 counter on this creature",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
