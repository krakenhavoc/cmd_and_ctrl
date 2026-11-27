package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Seismic Assault — Enchantment {R}{R}{R} (Eighth Edition, S82 OLD batch 610,
// #5607):
//
//	"Discard a land card: This enchantment deals 2 damage to any target."
//
// Molten Vortex's ability without the mana: a land card discarded as the cost,
// any number of times. The target is chosen as the ability is activated and
// re-checked as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8ad4f2fe-6d98-4279-a331-3817d40ae46d",
		Name:         "Seismic Assault",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Discard a land card: This enchantment deals 2 damage to any target.",
			Cost:    effects.DiscardCardsMatching(1, "a land card", game.Card.IsLand),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
			Effect:  effects.SourceDealsDamageToEachLegalTarget(2),
		}},
	})
}
