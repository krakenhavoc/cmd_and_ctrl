package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Way of the Wildspeaker — Legendary Enchantment {4}{G} (Reality
// Fracture, tracker #2795):
//
//	"When Way of the Wildspeaker enters, empower Jace 7.
//	 Planeswalkers you control have "[−4]: Create a 4/4 green Beast
//	 creature token with trample.""
//
// Empower Jace is the keyword action (ADR 0139); the granted row is an
// ADR 0093 bundle with a loyalty cost (ADR 0140), so the walker pays the
// four, and shares its one loyalty activation a turn with its other
// rows. The Beast is the activator's.
//
// No simplification.
func init() {
	const grant = "way-of-the-wildspeaker/beast"
	Register(Spec{
		OracleID:     "501580b9-692c-4804-be34-012a25ba8adf",
		Name:         "Way of the Wildspeaker",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Way of the Wildspeaker — empower Jace 7", Do(EmpowerJace{N: 7})),
		},
		Grants: []AbilityGrant{{
			Key: grant,
			Activated: []ActivatedAbility{{
				Label:   "−4: Create a 4/4 green Beast creature token with trample.",
				Cost:    LoyaltyCost(-4),
				Purpose: game.Purpose{Tokens: 1},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateToken{
						Controller: item.Controller,
						Template:   TokenCard("4/4 green Beast with trample"),
						N:          1,
					}.Apply(NewContext(g, item))
				},
			}},
			Text: "[−4]: Create a 4/4 green Beast creature token with trample.",
		}},
		Static: []game.StaticAbility{GrantAbilitiesToYourPlaneswalkers(grant)},
	})
}
