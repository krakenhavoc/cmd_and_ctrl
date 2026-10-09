package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sanctum Lurker — Creature — Horror {2}{B}, 3/2 (Reality Fracture,
// tracker #2795):
//
//	"When this creature enters, empower Jace 1.
//	 Planeswalkers you control aren't put into their owners' graveyards
//	 for having 0 loyalty.
//	 Planeswalkers you control have "[+2]: This planeswalker deals 1
//	 damage to each opponent and you gain 1 life.""
//
// Empower Jace is the keyword action (ADR 0139), and the token it makes
// sits at 1 loyalty. The exemption is ADR 0140's CR 704.5i static, read
// live, so a Lurker that loses its abilities exempts nothing. The granted
// +2 is an ADR 0093 bundle: the walker's own row, the damage dealt by the
// walker.
//
// No simplification.
func init() {
	const grant = "sanctum-lurker/drain"
	Register(Spec{
		OracleID:              "3ebd64d2-c178-45a2-a51f-9869754baa0f",
		Name:                  "Sanctum Lurker",
		Completeness:          CompletenessFull,
		ZeroLoyaltyExemptions: PlaneswalkersSurviveZeroLoyalty(),
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Sanctum Lurker — empower Jace 1", Do(EmpowerJace{N: 1})),
		},
		Grants: []AbilityGrant{{
			Key: grant,
			Activated: []ActivatedAbility{{
				Label:  "+2: This planeswalker deals 1 damage to each opponent and you gain 1 life.",
				Cost:   LoyaltyCost(2),
				Effect: damageEachOpponentThenGainLife(1),
			}},
			Text: "[+2]: This planeswalker deals 1 damage to each opponent and you gain 1 life.",
		}},
		Static: []game.StaticAbility{GrantAbilitiesToYourPlaneswalkers(grant)},
	})
}
