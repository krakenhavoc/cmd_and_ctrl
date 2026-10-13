package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rhystic Circle — Enchantment {2}{W}{W}:
//
//	"{1}: Any player may pay {1}. If no one does, the next time a
//	 source of your choice would deal damage to you this turn, prevent
//	 that damage."
//
// ADR 0150 (#1903): the {1} is offered to every player in APNAP order
// from the active player (AnyPlayerPays). If nobody pays, the shield
// is ADR 0107 §6's one-use shield against a chosen source (CR 615.8),
// the source chosen then (CR 609.7a), as on the Circles of Protection.
//
// No simplification.
func init() {
	shield := PreventNextDamageFromChosenSource(ShieldYou)
	shield.Question = "Rhystic Circle — choose a source"
	shield.Label = "Rhystic Circle — prevent the next damage from a source of your choice"
	Register(Spec{
		OracleID:     "bffb7d30-b9a4-4529-a3fc-c82d6cd4b599",
		Name:         "Rhystic Circle",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}: Any player may pay {1}. If no one does, the next time a source of your choice would deal damage to you this turn, prevent that damage.",
			Cost:    ManaCost("{1}"),
			Purpose: preventsDamageUnlessTargeted(nil),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return AnyPlayerPays{
					Cost:     "{1}",
					Question: "Rhystic Circle — pay {1} to stop the damage shield?",
					OnNobody: shield.Apply,
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
