package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Caustic Wasps — Creature — Insect {2}{G}, 1/1 (Mercadian Masques, S82 OLD,
// #5619):
//
//	"Flying
//	 Whenever this creature deals combat damage to a player, you may destroy
//	 target artifact that player controls."
//
// Flying is printed data. The target is chosen as the trigger goes on the stack,
// among the artifacts the damaged player controls (the player is read off the
// triggering event), and "you may" is asked as it resolves; the artifact is
// checked again then (CR 608.2b). A damaged player with no artifact leaves the
// trigger without a target, so it is removed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fea997a4-3ed1-4f83-b16a-56939f4b2211",
		Name:            "Caustic Wasps",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(effects.WheneverThisDealsCombatDamageToAPlayer("Caustic Wasps — you may destroy target artifact that player controls",
				func(g *game.Game, item *game.StackItem) error {
					return effects.DestroyTheTargetPermanent(item, effects.NewContext(g, item))
				}), effects.ControlledByThatPlayer(effects.TargetPermanent("target artifact that player controls", effects.Artifact()))),
				"Caustic Wasps — destroy target artifact that player controls?"),
		},
	})
}
