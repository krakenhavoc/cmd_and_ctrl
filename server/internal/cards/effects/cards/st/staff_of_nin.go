package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Staff of Nin — Artifact {6} (Magic 2013, #5568):
//
//	"At the beginning of your upkeep, draw a card.
//	 {T}: This artifact deals 1 damage to any target."
//
// An upkeep draw and Granite Shard's ping with a bare {T} cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "48d9f488-4ebb-4fdb-a568-6930fc2f5416",
		Name:         "Staff of Nin",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Staff of Nin — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: This artifact deals 1 damage to any target.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Cost:    effects.TapCost(),
			Targets: effects.TargetAny(),
			Effect:  effects.DealDamageToTheTarget(1),
		}},
	})
}
