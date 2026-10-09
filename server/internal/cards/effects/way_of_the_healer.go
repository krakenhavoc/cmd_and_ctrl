package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Way of the Healer — Legendary Enchantment {3}{W} (Reality Fracture,
// tracker #2795):
//
//	"When Way of the Healer enters, empower Jace 5.
//	 Planeswalkers you control have "[−2]: Create a 2/2 colorless Wizard
//	 Soldier creature token named Cadet. Surveil 1.""
//
// Empower Jace is the keyword action (ADR 0139); the granted row is an
// ADR 0093 bundle with a loyalty cost (ADR 0140). The surveil rides
// Surveil's own prompt, after the Cadet exists.
//
// No simplification.
func init() {
	const grant = "way-of-the-healer/cadet"
	Register(Spec{
		OracleID:     "cb27dc83-85ee-4f54-b032-41cd1806d3ac",
		Name:         "Way of the Healer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Way of the Healer — empower Jace 5", Do(EmpowerJace{N: 5})),
		},
		Grants: []AbilityGrant{{
			Key: grant,
			Activated: []ActivatedAbility{{
				Label:   "−2: Create a 2/2 colorless Wizard Soldier creature token named Cadet. Surveil 1.",
				Cost:    LoyaltyCost(-2),
				Purpose: game.Purpose{Tokens: 1},
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := (CreateToken{
						Controller: item.Controller,
						Template:   TokenCard("2/2 colorless Wizard Soldier named Cadet"),
						N:          1,
					}).Apply(ctx); err != nil {
						return err
					}
					return Surveil{Player: item.Controller, N: 1}.Apply(ctx)
				},
			}},
			Text: "[−2]: Create a 2/2 colorless Wizard Soldier creature token named Cadet. Surveil 1.",
		}},
		Static: []game.StaticAbility{GrantAbilitiesToYourPlaneswalkers(grant)},
	})
}
