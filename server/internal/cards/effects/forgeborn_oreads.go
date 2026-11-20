package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "8289190c-b180-48e3-b92f-7dccbca61b39",
		Name:         "Forgeborn Oreads",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(
				Targeting(On(game.EventETB, constellationTriggers, "Forgeborn Oreads — 1 damage to any target (constellation)",
					DealDamageToTheTarget(1)), TargetAny()),
				ForTargets(DamageToTarget(0, 1))),
		},
	})
}
