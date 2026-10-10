package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Terraformer — Creature — Human Wizard {2}{U}, 2/2:
//
//	"{1}: Choose a basic land type. Each land you control becomes that
//	 type until end of turn."
//
// ADR 0109 §1 (#1881): CR 305.7 over a set. The type is chosen as the
// ability resolves (CR 608.2), and the lands are the ones you control
// then: the set is fixed as the effect begins (CR 611.2c), so a land you
// play afterwards is not changed. Until end of turn each one's land types
// are replaced by the chosen one (other subtypes stay, CR 205.1a), it
// loses the abilities its rules text gives it, and it taps for the chosen
// colour (CR 305.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "683089a8-d5a9-438a-a463-77cc6148125e",
		Name:         "Terraformer",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}: Choose a basic land type. Each land you control becomes that type until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    ManaCost("{1}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return eachLandYouControlBecomesChosenTypeUntilEOT(NewContext(g, item), "Terraformer")
			},
		}},
	})
}
