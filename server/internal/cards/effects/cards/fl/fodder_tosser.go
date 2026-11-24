package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fodder Tosser — Artifact {3} (Modern Horizons 2):
//
//	"{T}, Discard a card: This artifact deals 2 damage to target player or
//	 planeswalker."
//
// The tap and the discard are costs; the target is a player or a planeswalker,
// not a creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "feb34b5a-d51e-42d7-90c6-c5c84ff00c4e",
		Name:         "Fodder Tosser",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Discard a card: This artifact deals 2 damage to target player or planeswalker.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
			Cost:    effects.Plus(effects.TapCost(), effects.DiscardACard()),
			Targets: effects.TargetPlayerOrPlaneswalker(),
			Effect:  effects.SourceDealsDamageToEachLegalTarget(2),
		}},
	})
}
