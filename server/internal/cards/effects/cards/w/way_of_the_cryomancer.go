package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Way of the Cryomancer — Legendary Enchantment {2}{U} (Reality
// Fracture, tracker #2795):
//
//	"When Way of the Cryomancer enters, empower Jace 5.
//	 Planeswalkers you control have "[−3]: When you next cast an instant
//	 or sorcery spell this turn, copy that spell. You may choose new
//	 targets for the copy.""
//
// Empower Jace is the keyword action (ADR 0139); the granted row is an
// ADR 0093 bundle with a loyalty cost (ADR 0140), and its body is
// Doublecast's: an event-conditioned delayed trigger that waits for the
// next instant or sorcery cast and ends with the turn.
//
// No simplification.
func init() {
	const grant = "way-of-the-cryomancer/copy"
	effects.Register(effects.Spec{
		OracleID:     "7d32c63d-aa9e-4f5d-a3e6-51377a786009",
		Name:         "Way of the Cryomancer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Way of the Cryomancer — empower Jace 5", effects.Do(effects.EmpowerJace{N: 5})),
		},
		Grants: []effects.AbilityGrant{{
			Key: grant,
			Activated: []effects.ActivatedAbility{{
				Label: "−3: When you next cast an instant or sorcery spell this turn, copy that spell. You may choose new targets for the copy.",
				Cost:  effects.LoyaltyCost(-3),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.WhenYouNextCast(
						"Way of the Cryomancer — copy that spell",
						game.CastFilter{Types: []string{"Instant", "Sorcery"}},
						effects.CopyTheSpellBody,
					).Apply(effects.NewContext(g, item))
				},
			}},
			Text: "[−3]: When you next cast an instant or sorcery spell this turn, copy that spell. You may choose new targets for the copy.",
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToYourPlaneswalkers(grant)},
	})
}
