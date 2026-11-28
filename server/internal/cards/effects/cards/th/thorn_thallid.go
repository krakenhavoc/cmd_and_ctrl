package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thorn Thallid — Creature — Fungus {1}{G}{G}, 2/2 (Fallen Empires, S82 OLD
// b853, #5637):
//
//	"At the beginning of your upkeep, put a spore counter on this creature.
//	 Remove three spore counters from this creature: It deals 1 damage to any
//	 target."
//
// The damage is dealt by the Thallid, read as it last existed if it has gone
// by the time the ability resolves (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1239d3bf-06e8-4732-b1e2-f6eb145e5894",
		Name:         "Thorn Thallid",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.S82OldB853SporeUpkeep("Thorn Thallid")},
		Activated: []effects.ActivatedAbility{{
			Label:   "Remove three spore counters from this creature: It deals 1 damage to any target.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Cost:    effects.S82OldB853SporeCost(),
			Targets: effects.TargetAny(),
			Effect:  effects.DamageToFirstTargetEffect(1),
		}},
	})
}
