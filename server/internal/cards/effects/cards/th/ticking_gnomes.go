package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ticking Gnomes — Artifact Creature — Gnome, {3}, 3/3:
//
//	"Echo {3} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	 Sacrifice this creature: It deals 1 damage to any target."
//
// The sacrifice is part of the cost, so the damage comes from the Gnomes as
// they last existed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "963de959-848d-48df-90fa-88a6d988b9fe",
		Name:         "Ticking Gnomes",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this creature: It deals 1 damage to any target.",
			Cost:    effects.SacrificeThis(),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Effect:  effects.SourceDealsOneToFirstTarget,
		}},
		Triggered: []game.TriggeredAbility{effects.Echo("Ticking Gnomes", "{3}")},
	})
}
