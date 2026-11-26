package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blizzard Strix — Snow Creature — Bird {4}{U}, 3/2 (Modern Horizons):
//
//	"Flash
//	 Flying
//	 When this creature enters, if you control another snow permanent, exile
//	 target permanent other than this creature. Return that card to the
//	 battlefield under its owner's control at the beginning of the next end
//	 step."
//
// The trigger has an intervening "if" (CR 603.4), read as the Strix enters and
// again as the ability resolves; with no other snow permanent it never goes on
// the stack. The target (any permanent but the Strix itself) is chosen as the
// ability is put on the stack and is exiled as it resolves, and the one delayed
// trigger returns the card that reached exile at the next end step under its
// owner's control (CR 603.7), tokens never coming back (CR 111.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c687af9b-2159-4965-a068-ce4db1ce801c",
		Name:            "Blizzard Strix",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.OldB166IfYouControlAnotherSnowPermanent(
				effects.WhenThisEnters("Blizzard Strix — exile target permanent other than this creature; return it at the beginning of the next end step",
					func(g *game.Game, item *game.StackItem) error {
						return effects.ExileTargetsThenScheduleReturn(effects.NewContext(g, item), "Blizzard Strix — return the exiled card")
					})),
				effects.Another(effects.TargetPermanent("target permanent other than this creature"))),
		},
	})
}
