package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rustmouth Ogre — Creature — Ogre {4}{R}{R}, 5/4 (Mirrodin, S82 OLD,
// #5606):
//
//	"Whenever this creature deals combat damage to a player, you may
//	 destroy target artifact that player controls."
//
// Blind Zealot's shape: the target is chosen as the trigger goes on the
// stack, among the artifacts the damaged player controls (an artifact
// creature included), and the "you may" is asked as it resolves (CR 603.5).
// An artifact that has left or has stopped being legal is not destroyed
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0a64b683-7017-44cb-923f-b512fe3886f6",
		Name:         "Rustmouth Ogre",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(effects.WheneverThisDealsCombatDamageToAPlayer(
				"Rustmouth Ogre — you may destroy target artifact that player controls", effects.DestroyFirstLegalTarget),
				effects.ControlledByThatPlayer(effects.TargetPermanent("target artifact that player controls", effects.Artifact()))),
				"Rustmouth Ogre — destroy target artifact that player controls?"),
		},
	})
}
