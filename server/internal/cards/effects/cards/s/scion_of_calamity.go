package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "75fddffb-68f4-4b65-bf3b-4e2ada62a522",
		Name:         "Scion of Calamity",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.GrantedMyriad("Scion of Calamity"),
			effects.Targeting(effects.WheneverThisDealsCombatDamageToAPlayer("Scion of Calamity — destroy target artifact or enchantment that player controls",
				effects.DestroyFirstLegalTarget),
				effects.ControlledByThatPlayer(effects.TargetPermanent("target artifact or enchantment that player controls",
					effects.Or(effects.Artifact(), effects.Enchantment())))),
		},
	})
}
