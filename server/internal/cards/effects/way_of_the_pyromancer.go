package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Way of the Pyromancer — Legendary Enchantment {1}{R} (Reality
// Fracture, tracker #2795):
//
//	"When Way of the Pyromancer enters, empower Jace 2.
//	 Planeswalkers you control have "[+1]: Add {R}.""
//
// Empower Jace is the keyword action (ADR 0139); the granted row is an
// ADR 0093 bundle with a loyalty cost (ADR 0140). A loyalty ability is
// never a mana ability (CR 605.1a), so the +1 goes on the stack and can
// be responded to; AddMana is the stack-using form (CR 605.3a's
// counterpart). The red mana empties with the step (CR 106.4).
//
// No simplification.
func init() {
	const grant = "way-of-the-pyromancer/mana"
	Register(Spec{
		OracleID:     "68d3f547-674a-4d8a-b02f-01bc2f161916",
		Name:         "Way of the Pyromancer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Way of the Pyromancer — empower Jace 2", Do(EmpowerJace{N: 2})),
		},
		Grants: []AbilityGrant{{
			Key: grant,
			Activated: []ActivatedAbility{{
				Label: "+1: Add {R}.",
				Cost:  LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return AddMana{Produced: "{R}"}.Apply(NewContext(g, item))
				},
			}},
			Text: "[+1]: Add {R}.",
		}},
		Static: []game.StaticAbility{GrantAbilitiesToYourPlaneswalkers(grant)},
	})
}
