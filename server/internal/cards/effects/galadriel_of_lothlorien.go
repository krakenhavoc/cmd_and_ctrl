package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Galadriel of Lothlórien — Legendary Creature — Elf Noble
// {1}{G}{U}, 3/3:
//
//	"Whenever the Ring tempts you, if you chose a creature other than
//	 Galadriel as your Ring-bearer, scry 3.
//	 Whenever you scry, you may reveal the top card of your library.
//	 If a land card is revealed this way, put it onto the battlefield
//	 tapped."
//
// The second ability watches EventScry, which is emitted once a scry
// is complete: "You finish scrying before revealing the top card of
// your library" (2023-06-16 ruling). A scry that looked at nothing (an
// empty library) is no scry and does not trigger it. It triggers on
// every scry you perform, the first ability's included. The "you may"
// is asked as it resolves (CR 608.2d).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8d3c9646-2b6b-4315-939d-5a920efc0a60",
		Name:         "Galadriel of Lothlórien",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			IfYouChoseAnotherRingBearer(WheneverTheRingTemptsYou("Galadriel of Lothlórien — scry 3", Do(Scry{N: 3}))),
			On(game.EventScry, ByYou, "Galadriel of Lothlórien — you may reveal the top card of your library", galadrielOfLothlorienReveal),
		},
	})
}

// galadrielOfLothlorienReveal is the second ability's body: the "you
// may", then the reveal and the land it may put onto the battlefield.
func galadrielOfLothlorienReveal(g *game.Game, item *game.StackItem) error {
	return MayChoice{
		Question: "Galadriel of Lothlórien — reveal the top card of your library?",
		OnYes: func(ctx *Context) error {
			_, err := revealTopThenPutIfMatch(ctx.Game, ctx.Source(), ctx.Controller(), game.Card.IsLand, true,
				"Galadriel of Lothlórien — revealed from the top of the library")
			return err
		},
	}.Apply(NewContext(g, item))
}
