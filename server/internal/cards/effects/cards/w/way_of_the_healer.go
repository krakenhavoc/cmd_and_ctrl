package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "cb27dc83-85ee-4f54-b032-41cd1806d3ac",
		Name:         "Way of the Healer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Way of the Healer — empower Jace 5", effects.Do(effects.EmpowerJace{N: 5})),
		},
		Grants: []effects.AbilityGrant{{
			Key: grant,
			Activated: []effects.ActivatedAbility{{
				Label:   "−2: Create a 2/2 colorless Wizard Soldier creature token named Cadet. Surveil 1.",
				Cost:    effects.LoyaltyCost(-2),
				Purpose: game.Purpose{Tokens: 1},
				// The surveil waits for a prompt the token asks (a CR 616 order; CR 608.2c, #4155).
				Effect: effects.Do(
					effects.CreateToken{
						Template: effects.TokenCard("2/2 colorless Wizard Soldier named Cadet"),
						N:        1,
					},
					effects.Surveil{N: 1},
				),
			}},
			Text: "[−2]: Create a 2/2 colorless Wizard Soldier creature token named Cadet. Surveil 1.",
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToYourPlaneswalkers(grant)},
	})
}
