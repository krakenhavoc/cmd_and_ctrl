package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Forgeborn Oreads — Enchantment Creature — Nymph {2}{R}{R}, 4/2 (Journey into Nyx, #5558):
//
//	"Constellation — Whenever this creature or another enchantment you
//	 control enters, this creature deals 1 damage to any target."
//
// Constellation is Eidolon of Blossoms' condition: its own entry (it is an
// enchantment) or another enchantment entering under your control. The
// ping is Anaba Shaman's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8289190c-b180-48e3-b92f-7dccbca61b39",
		Name:         "Forgeborn Oreads",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(
				effects.Targeting(effects.On(game.EventETB, effects.ConstellationTriggers, "Forgeborn Oreads — 1 damage to any target (constellation)",
					effects.DealDamageToTheTarget(1)), effects.TargetAny()),
				effects.ForTargets(effects.DamageToTarget(0, 1))),
		},
	})
}
