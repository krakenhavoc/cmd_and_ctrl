package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mistmeadow Vanisher — Creature — Kithkin Wizard {2}{W/U}, 3/2 (March of
// the Machine Commander, #5477):
//
//	"Whenever this creature becomes tapped, exile up to one target
//	 nonland, nontoken permanent. Return that card to the battlefield
//	 under its owner's control at the beginning of the next end step."
//
// Attentive Sunscribe's "becomes tapped" event (ThisBecameTapped: the tap
// event and the attack that taps it, so a vigilance attack is not one)
// with Skyskipper Duo's delayed blink. The target may be any player's
// permanent, or the Vanisher itself. A token is not a legal target, so
// nothing exiled here is ever lost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5d11ada1-6459-4a91-bd10-66673dc1ad3c",
		Name:         "Mistmeadow Vanisher",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, effects.ThisBecameTapped,
					"Mistmeadow Vanisher — exile up to one target nonland, nontoken permanent; return it at the beginning of the next end step",
					func(g *game.Game, item *game.StackItem) error {
						return effects.ExileTargetsThenScheduleReturn(effects.NewContext(g, item), "Mistmeadow Vanisher — return the exiled card")
					}),
				effects.TargetPermanent("up to one target nonland, nontoken permanent", effects.Nonland(), effects.Not(effects.IsTokenPredicate())).WithCount(0, 1),
			),
		},
	})
}
