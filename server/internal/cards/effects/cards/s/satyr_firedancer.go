package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Satyr Firedancer — Enchantment Creature — Satyr {1}{R}, 1/1 (Born of the
// Gods, S82 OLD b378, #5559):
//
//	"Whenever an instant or sorcery spell you control deals damage to an
//	 opponent, this creature deals that much damage to target creature that
//	 player controls."
//
// Brash Taunter's batch total (ThatMuchDamage: one trigger per damage
// instance, the sum of what the spell dealt) on Bloodfeather Phoenix's
// "instant or sorcery spell you control" read mid-resolution. The target is a
// creature the damaged player controls (ControlledByThatPlayer).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b30edff0-3716-4dbb-b445-5329b08e36f4",
		Name:         "Satyr Firedancer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:    []game.EventKind{game.EventDealDamage},
			AppliesTo:  effects.S82OldB378SpellDamagedAnOpponent,
			Targets:    effects.ControlledByThatPlayer(effects.TargetCreature("target creature that player controls")),
			Key:        "Satyr Firedancer — it deals that much damage to target creature that player controls",
			AtBatchEnd: effects.ThatMuchDamage(),
			Effect:     effects.S82OldB378ThatMuchToTarget,
		}},
	})
}
