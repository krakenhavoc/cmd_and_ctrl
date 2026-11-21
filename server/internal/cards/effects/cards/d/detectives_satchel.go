package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "e4ecfccd-b8b4-4b18-b989-30de79b4971e",
		Name:         "Detective's Satchel",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Detective's Satchel — investigate twice", effects.Do(effects.Investigate{N: 2})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:     "{T}: Create a 1/1 colorless Thopter artifact creature token with flying. Activate only if you've sacrificed an artifact this turn.",
			Cost:      effects.TapCost(),
			Condition: effects.YouSacrificedAnArtifactThisTurn,
			Purpose:   game.Purpose{Answers: game.AnswerMakesBlocker, Tokens: 1},
			Effect:    effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 colorless Thopter artifact with flying"), N: 1}),
		}},
	})
}
