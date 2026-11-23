package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	shield := effects.PreventNextDamageFromChosenSource(effects.ShieldYou)
	shield.Question = "Rhystic Circle — choose a source"
	shield.Label = "Rhystic Circle — prevent the next damage from a source of your choice"
	// "If no one does" is keyed (#4264): the shield is built here and the
	// key names it, so a table waiting on any player's answer is a
	// restore point.
	rhysticCircleShield := effects.AnyPlayerConsequences("any-player/rhystic-circle-shield", nil,
		func(ctx *effects.Context, _ game.AnyPlayerAnswer) error { return shield.Apply(ctx) }, nil)
	effects.Register(effects.Spec{
		OracleID:     "bffb7d30-b9a4-4529-a3fc-c82d6cd4b599",
		Name:         "Rhystic Circle",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}: Any player may pay {1}. If no one does, the next time a source of your choice would deal damage to you this turn, prevent that damage.",
			Cost:    effects.ManaCost("{1}"),
			Purpose: effects.PreventsDamage(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.AnyPlayerPays{
					Cost:     "{1}",
					Question: "Rhystic Circle — pay {1} to stop the damage shield?",
					Then:     rhysticCircleShield,
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
