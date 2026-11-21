package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Overlord of the Boilerbilges — Enchantment Creature — Avatar Horror
// {4}{R}{R}, 5/5 (Duskmourn: House of Horror, #4466):
//
//	"Impending 4—{2}{R}{R} (If you cast this spell for its impending
//	 cost, it enters with four time counters and isn't a creature
//	 until the last is removed. At the beginning of your end step,
//	 remove a time counter from it.)
//	 Whenever this permanent enters or attacks, it deals 4 damage to
//	 any target."
//
// Impending is the shared constructor (impending.go); the card adds the
// trigger, aimed as it goes on the stack and dealing its 4 only if the
// target is still legal as it resolves (DealDamageToTheTarget, CR
// 608.2b). Impended, it is not a creature and cannot attack, but the
// enter half still fires.
//
// No simplification.
func init() {
	effects.Register(effects.Impending(4, "{2}{R}{R}", effects.Spec{
		OracleID:     "c2db05e0-879b-43e4-9009-37a95fd848dc",
		Name:         "Overlord of the Boilerbilges",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(
				effects.WhenThisEntersOrAttacks("Overlord of the Boilerbilges — 4 damage to any target", effects.DealDamageToTheTarget(4)),
				effects.TargetAny(),
			), effects.ForTargets(effects.DamageToTarget(0, 4))),
		},
	}))
}
