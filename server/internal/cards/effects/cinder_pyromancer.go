package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "47bfee3e-34b4-4b89-aaf8-f74e0c3d3c62",
		Name:         "Cinder Pyromancer",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: This creature deals 1 damage to target player or planeswalker.",
			Cost:    TapCost(),
			Targets: targetPlayerOrPlaneswalker(),
			Purpose: ForTargets(DamageToTarget(0, 1)),
			Effect:  sourceDealsOneToFirstTarget,
		}},
		Triggered: []game.TriggeredAbility{eveB1UntapOnCast("Cinder Pyromancer", "R")},
	})
}
