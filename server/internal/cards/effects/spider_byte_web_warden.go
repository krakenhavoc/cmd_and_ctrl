package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spider-Byte, Web Warden — Legendary Creature — Spider Avatar Hero {2}{U}, 2/2
// (Marvel's Spider-Man, #4460):
//
//	"When Spider-Byte enters, return up to one target nonland permanent to its
//	 owner's hand."
//
// Star Athlete's "up to one target nonland permanent" clause with the bounce
// body Repulsor Bots uses (b35BounceChosenPermanents). Choosing no target is
// legal and returns nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "845472ca-3fcb-4bcd-953c-96eb8d914af9",
		Name:         "Spider-Byte, Web Warden",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Spider-Byte — return up to one target nonland permanent to its owner's hand",
				func(g *game.Game, item *game.StackItem) error {
					return b35BounceChosenPermanents(NewContext(g, item))
				}),
				TargetPermanent("up to one target nonland permanent", Nonland()).WithCount(0, 1)),
		},
	})
}
