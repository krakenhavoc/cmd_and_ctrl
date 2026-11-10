package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Detective's Satchel — Artifact {2}{U}{R} (Murders at Karlov Manor,
// #4469):
//
//	"When this artifact enters, investigate twice. (To investigate,
//	 create a Clue token. It's an artifact with "{2}, Sacrifice this
//	 token: Draw a card.")
//	 {T}: Create a 1/1 colorless Thopter artifact creature token with
//	 flying. Activate only if you've sacrificed an artifact this turn."
//
// Investigating twice creates two Clue tokens (CR 701.16a). The
// activation condition reads the turn's card-type sacrifice tally
// (YouSacrificedAnArtifactThisTurn, #5224), so a Clue cracked this turn
// counts though the token no longer exists, and so does any other
// artifact, an artifact creature included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e4ecfccd-b8b4-4b18-b989-30de79b4971e",
		Name:         "Detective's Satchel",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Detective's Satchel — investigate twice", Do(Investigate{N: 2})),
		},
		Activated: []ActivatedAbility{{
			Label:     "{T}: Create a 1/1 colorless Thopter artifact creature token with flying. Activate only if you've sacrificed an artifact this turn.",
			Cost:      TapCost(),
			Condition: YouSacrificedAnArtifactThisTurn,
			Purpose:   game.Purpose{Answers: game.AnswerMakesBlocker, Tokens: 1},
			Effect:    Do(CreateToken{Template: TokenCard("1/1 colorless Thopter artifact with flying"), N: 1}),
		}},
	})
}
