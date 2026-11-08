package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shady Informant — Creature — Ogre Rogue {3}{B}{R}, 4/2 (Murders at
// Karlov Manor, #4469):
//
//	"When this creature dies, it deals 2 damage to any target.
//	 Disguise {2}{B/R}{B/R} (You may cast this card face down for {3} as
//	 a 2/2 creature with ward {2}. Turn it face up any time for its
//	 disguise cost.)"
//
// Dynamite Diver's death trigger with 2 damage, dealt by the Informant
// as it last existed on the battlefield (CR 608.2h). A face-down
// Informant has no abilities (CR 708.2a), so one that dies face down
// deals nothing. Disguise is Alley Assailant's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "5cf3727b-eb7b-414e-982c-3acf4c05ddf1",
		Name:             "Shady Informant",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Disguise("{2}{B/R}{B/R}")},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WhenThisDies("Shady Informant — 2 damage to any target",
				sourceDealsDamageToEachLegalTarget(2)), TargetAny()), ForTargets(DamageToTarget(0, 2))),
		},
	})
}
