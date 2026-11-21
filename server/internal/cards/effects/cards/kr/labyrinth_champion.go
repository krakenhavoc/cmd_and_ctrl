package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Labyrinth Champion — Creature — Human Warrior {3}{R}, 2/2 (Journey into
// Nyx):
//
//	"Heroic — Whenever you cast a spell that targets this creature, this
//	 creature deals 2 damage to any target."
//
// Heroic is an ability word with no rules of its own: the trigger is the shared
// Heroic constructor (a spell you cast that targets it, once per spell). Its target
// is chosen as the trigger goes on the stack and re-checked as it resolves
// (CR 608.2b); the damage is dealt by the Champion as it last existed (CR
// 608.2h) if it has left.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e1a22f41-0543-4e85-8fb8-93c018b461aa",
		Name:         "Labyrinth Champion",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(
				effects.Heroic("Labyrinth Champion — 2 damage to any target", effects.SourceDealsDamageToEachLegalTarget(2)),
				effects.TargetAny()), effects.ForTargets(effects.DamageToTarget(0, 2))),
		},
	})
}
