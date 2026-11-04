package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Donatello, Gadget Master — Legendary Creature — Mutant Ninja Turtle
// {2}{U}, 3/2 (TMT, #4457):
//
//	"Sneak {1}{U} (… He enters tapped and attacking.)
//	 Whenever Donatello deals combat damage to a player, create a token
//	 that's a copy of target artifact you control."
//
// Sneak is Sneak(cost) (CR 702.190). The trigger is Artificer Class's
// level-3 body behind a combat-damage condition: the token copies the
// artifact's copiable values (CR 707.2), and a target that left in
// response makes nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "bb6aef0d-052b-44e6-b4cd-2d2eafd84969",
		Name:             "Donatello, Gadget Master",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Sneak("{1}{U}")},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverThisDealsCombatDamageToAPlayer("Donatello, Gadget Master — create a token that's a copy of target artifact you control",
				func(g *game.Game, item *game.StackItem) error {
					return TokenCopyOfSingleTarget(item, NewContext(g, item))
				}),
				TargetPermanent("target artifact you control", Artifact(), YouControl())),
		},
	})
}
