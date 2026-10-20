package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chromatic Sphere — Artifact {1}:
//
//	"{1}, {T}, Sacrifice this artifact: Add one mana of any color.
//	 Draw a card. (Activate only as an instant.)"
//
// Chromatic Star's sibling with the draw on the ability instead of on
// leaving the battlefield. An ordinary activated ability, not a mana
// ability (#3971): its effect draws a card, which moves a card from a
// library, and CR 605.1a excludes that. So it uses the stack: the {1},
// the {T} and the sacrifice are paid at announce (CR 602.2b), an
// opponent may respond (the Sphere is already gone), and on resolution
// the controller adds one mana of any color and then draws a card, in
// the printed order. "Activate only as an instant" is every activated
// ability's default timing. The auto-tapper never plans it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2e03e44a-9fff-4490-859f-b42e89e8563a",
		Name:         "Chromatic Sphere",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}, Sacrifice this artifact: Add one mana of any color. Draw a card",
			Cost:    Plus(ManaCost("{1}"), TapCost(), SacrificeThis()),
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1}, // ADR 0142
			Effect:  Do(AddMana{Produced: "{W|U|B|R|G}"}, DrawCards{N: 1}),
		}},
	})
}
