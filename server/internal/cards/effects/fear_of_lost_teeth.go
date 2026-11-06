package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fear of Lost Teeth — Enchantment Creature — Nightmare {B}, 1/1
// (Duskmourn: House of Horror, #4466):
//
//	"When this creature dies, it deals 1 damage to any target and you
//	 gain 1 life."
//
// Dynamite Diver's death trigger, then the life: the dead creature is the
// damage's source, and its controller gains the life. With its only
// target gone by resolution the trigger does nothing, the life included
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3724cc42-1bbd-4a11-8469-ab8492ddba52",
		Name:         "Fear of Lost Teeth",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WhenThisDies("Fear of Lost Teeth — 1 damage to any target and you gain 1 life",
				func(g *game.Game, item *game.StackItem) error {
					if err := sourceDealsDamageToEachLegalTarget(1)(g, item); err != nil {
						return err
					}
					return GainLife{Amount: 1}.Apply(NewContext(g, item))
				}), TargetAny()), ForTargets(DamageToTarget(0, 1))),
		},
	})
}
