package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sunder Shaman — Creature — Giant Shaman {R}{R}{G}{G}, 5/5 (Ravnica
// Remastered, #5948):
//
//	"This creature can't be blocked by more than one creature.
//	 Whenever this creature deals combat damage to a player, destroy target
//	 artifact or enchantment that player controls."
//
// The block rule is Charging Rhino's (MaxBlockers on itself). The trigger's
// target is restricted to permanents the damaged player controls
// (ControlledByThatPlayer, Dreadmaw's Ire), judged when it is put on the
// stack and again as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "38b23fe1-0e76-494a-ad9b-bc22c7d21a5e",
		Name:         "Sunder Shaman",
		Completeness: CompletenessFull,
		BlockRules:   []game.BlockRule{MaxBlockers(OnSelf(), 1)},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WheneverThisDealsCombatDamageToAPlayer("Sunder Shaman — destroy target artifact or enchantment that player controls",
					destroyFirstLegalTarget),
				ControlledByThatPlayer(TargetPermanent("target artifact or enchantment that player controls", Or(Artifact(), Enchantment())))),
		},
	})
}
