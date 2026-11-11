package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scion of Calamity — Creature — Dinosaur {3}{G}{G}, 5/5 (Lost Caverns of
// Ixalan Commander):
//
//	"Myriad (...)
//	 Whenever this creature deals combat damage to a player, destroy target
//	 artifact or enchantment that player controls."
//
// Myriad is the shared trigger (myriad.go, CR 702.116): the creature's own,
// one "you may" per other opponent, the copies exiled at end of combat. The
// damage trigger is Trygon Predator's without the "may": the target is
// chosen among what the damaged player controls as the trigger goes on the
// stack and re-checked as it resolves. Each myriad copy is a Scion too.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "75fddffb-68f4-4b65-bf3b-4e2ada62a522",
		Name:         "Scion of Calamity",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			GrantedMyriad("Scion of Calamity"),
			Targeting(WheneverThisDealsCombatDamageToAPlayer("Scion of Calamity — destroy target artifact or enchantment that player controls",
				destroyFirstLegalTarget),
				ControlledByThatPlayer(TargetPermanent("target artifact or enchantment that player controls",
					Or(Artifact(), Enchantment())))),
		},
	})
}
