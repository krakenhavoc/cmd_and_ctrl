package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Maverick Thopterist — Creature — Human Artificer {3}{U}{R}, 2/2:
//
//	"Improvise ... When this creature enters, create two 1/1 colorless
//	 Thopter artifact creature tokens with flying."
//
// Improvise is Spec.TapCost. No simplification.
func init() {
	Register(Spec{
		OracleID:     "71ba6f1b-2fa6-4d4c-8778-f0765b1a5d8e",
		Name:         "Maverick Thopterist",
		Completeness: CompletenessFull,
		TapCost:      Improvise(),
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Maverick Thopterist — create two Thopters",
				Do(CreateToken{Template: TokenCard("1/1 colorless Thopter artifact with flying"), N: 2})),
		},
	})
}
