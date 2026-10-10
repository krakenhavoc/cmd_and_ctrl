package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Master's Guide-Mural // Master's Manufactory — a transforming artifact
// (#2709, ADR 0137):
//
//	Master's Guide-Mural — Artifact {3}{W}{U}
//	  "When this artifact enters, create a 4/4 white and blue Golem
//	   artifact creature token.
//	   Craft with artifact {4}{W}{W}{U}"
//	Master's Manufactory — Artifact
//	  "{T}: Create a 4/4 white and blue Golem artifact creature token.
//	   Activate only if this artifact or another artifact entered the
//	   battlefield under your control this turn."
//
// The Manufactory's condition reads the turn's entry tally by card
// type (game.EnteredWithCardTypeThisTurn), recorded as each permanent
// entered: per the ruling, that artifact may since have left, stopped
// being an artifact or changed controller. The crafted Manufactory
// itself entered this turn, so it can make a Golem the turn it is made.
//
// No simplification.
const mastersGuideMuralOracleID = "2dfdd358-e3cd-4bf5-871b-37689b7951e9"

func init() {
	Register(Spec{
		OracleID:     mastersGuideMuralOracleID,
		Name:         "Master's Guide-Mural",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Master's Guide-Mural — create a 4/4 Golem",
				Do(CreateToken{Template: TokenCard("4/4 white and blue Golem artifact"), N: 1})),
		},
		Activated: []ActivatedAbility{
			Craft("Craft with artifact {4}{W}{W}{U}", "{4}{W}{W}{U}", CraftWith("artifact")),
		},
	})

	Register(Spec{
		OracleID:     mastersGuideMuralOracleID + "#1",
		Name:         "Master's Manufactory",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     "{T}: Create a 4/4 white and blue Golem artifact creature token. Activate only if this artifact or another artifact entered the battlefield under your control this turn.",
			Cost:      TapCost(),
			Purpose:   game.Purpose{Answers: game.AnswerMakesBlocker, Tokens: 1},
			Condition: anArtifactEnteredUnderYouThisTurn,
			Effect: func(g *game.Game, item *game.StackItem) error {
				return CreateToken{Template: TokenCard("4/4 white and blue Golem artifact"), N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}

// anArtifactEnteredUnderYouThisTurn is "if this artifact or another
// artifact entered the battlefield under your control this turn".
func anArtifactEnteredUnderYouThisTurn(g *game.Game, controller, _ uuid.UUID) bool {
	return g.EnteredWithCardTypeThisTurn(controller, "artifact") > 0
}
