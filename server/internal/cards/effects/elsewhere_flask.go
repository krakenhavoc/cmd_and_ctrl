package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Elsewhere Flask — Artifact {2}:
//
//	"When this artifact enters, draw a card.
//	 Sacrifice this artifact: Choose a basic land type. Each land you
//	 control becomes that type until end of turn."
//
// ADR 0109 §1 (#1881): Terraformer's ability for a sacrifice. The type is
// chosen as the ability resolves (CR 608.2) and the lands are the ones you
// control then (CR 611.2c). Until end of turn each one's land types are
// replaced by the chosen one (other subtypes stay, CR 205.1a), it loses
// the abilities its rules text gives it, and it taps for the chosen colour
// (CR 305.6). The sacrifice is a cost, so it can be activated at instant
// speed with the Flask leaving at once.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "610618f9-8186-44b6-8041-0a5b70f11cc1",
		Name:         "Elsewhere Flask",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Elsewhere Flask — draw a card", Do(DrawCards{N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice this artifact: Choose a basic land type. Each land you control becomes that type until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    SacrificeThis(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return eachLandYouControlBecomesChosenTypeUntilEOT(NewContext(g, item), "Elsewhere Flask")
			},
		}},
	})
}
