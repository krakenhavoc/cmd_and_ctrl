package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Firestarter — Creature — Goblin {R}, 1/1 (#4062):
//
//	"Sacrifice this creature: It deals 1 damage to any target.
//	 Activate only during your turn, before attackers are declared."
//
// The window is DuringYourTurnBeforeAttackersAreDeclared, an activation
// condition (CR 602.1b): your own turn, and only until the first declare
// attackers step of the turn begins - the card's ruling reads it from the
// first combat, so a turn with several combat phases does not open it
// again. Not a sorcery-speed gate: it opens your beginning, combat and
// main-phase steps up to that point, with or without an item on the stack,
// and the view and the legal enumerator read the same closure, so the
// ability greys out when the window shuts.
//
// The sacrifice is a cost, so the damage is dealt by the creature as it
// last existed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "18279733-5568-427c-b0a2-044bca28f688",
		Name:         "Goblin Firestarter",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     "Sacrifice this creature: It deals 1 damage to any target. Activate only during your turn, before attackers are declared.",
			Cost:      effects.SacrificeThis(),
			Targets:   effects.TargetAny(),
			Purpose:   effects.ForTargets(effects.DamageToTarget(0, 1)),
			Condition: effects.DuringYourTurnBeforeAttackersAreDeclared(),
			Effect:    effects.B33DamageChosenTargetFromSource(1),
		}},
	})
}
