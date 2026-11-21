package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cinder Pyromancer — Creature — Elemental Shaman {2}{R}, 0/1 (Eventide, #5590):
//
//	"{T}: This creature deals 1 damage to target player or planeswalker.
//	 Whenever you cast a red spell, you may untap this creature."
//
// Goblin Fireslinger's ping and the optional cast-trigger untap
// (eveB1UntapOnCast).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "47bfee3e-34b4-4b89-aaf8-f74e0c3d3c62",
		Name:         "Cinder Pyromancer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: This creature deals 1 damage to target player or planeswalker.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetPlayerOrPlaneswalker(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Effect:  effects.SourceDealsOneToFirstTarget,
		}},
		Triggered: []game.TriggeredAbility{effects.EveB1UntapOnCast("Cinder Pyromancer", "R")},
	})
}
