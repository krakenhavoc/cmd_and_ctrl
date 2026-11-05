package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Astute Terraformer — Creature — Vulcan Scientist {1}{G}, 2/2 (Star Trek,
// #4973):
//
//	"Strange new worlds — {6}: This creature gets +2/+2 until end of turn.
//	 This ability costs {X} less to activate, where X is the number of
//	 differently named lands you control. Activate only once each turn."
//
// "Strange new worlds" is an ability word with no rules meaning (CR
// 207.2c). The discount is the ability's own cost clause (Suus Mahna
// Practitioner's shape), counted as the total cost is determined (CR
// 601.2f): each land the activator controls, once per name, off the
// generic {6}. "Activate only once each turn" gates the offer (CR 602.5b).
//
// No simplification.
func init() {
	const label = "Strange new worlds — {6}: This creature gets +2/+2 until end of turn. This ability costs {X} less to activate, where X is the number of differently named lands you control. Activate only once each turn."
	Register(Spec{
		OracleID:     "c5a7a82e-3c13-4fa1-8bf5-bb85b2890f75",
		Name:         "Astute Terraformer",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerPump},
			Cost:      ManaCost("{6}"),
			Condition: OncePerTurnActivation(label),
			CostModifiers: []game.CostModifier{
				CostsLessEach(trekB9DifferentlyNamedPermanentsYouControl(Land()),
					"This ability costs {X} less to activate, where X is the number of differently named lands you control."),
			},
			Effect: thisGetsUntilEndOfTurn(2, 2, "Astute Terraformer — +2/+2 until end of turn"),
		}},
	})
}
