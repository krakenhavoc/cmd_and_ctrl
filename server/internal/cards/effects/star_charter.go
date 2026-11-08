package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Star Charter — Creature — Bat Cleric {3}{W}, 3/1 (Bloomburrow, #4467):
//
//	"Flying
//	 At the beginning of your end step, if you gained or lost life this
//	 turn, look at the top four cards of your library. You may reveal a
//	 creature card with power 3 or less from among them and put it into
//	 your hand. Put the rest on the bottom of your library in a random
//	 order."
//
// "If you gained or lost life this turn" is an intervening if (CR 603.4)
// read off the per-turn life tallies as your end step begins. The rule
// asks again as the trigger resolves, but those tallies only climb
// within a turn, so a condition true at the trigger is still true then.
// The look is Dragonologist's LookAtTopThenMayTakeToHand: only you see
// the four, the card taken is revealed, and the rest go to the bottom in
// a random order. "Power 3 or less" is read off the card in the library,
// its printed power.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "578ddb51-e128-43e5-a115-471f0cae1222",
		Name:            "Star Charter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				you := source.Controller
				return ev.Actor == you && (b15LifeGainedThisTurn(g, you) > 0 || b18LifeLostThisTurn(g, you) > 0)
			}, "Star Charter — look at the top four cards; you may reveal a creature card with power 3 or less and put it into your hand",
				LookAtTopThenMayTakeToHand(4, And(Creature(), PowerLE(3)), 1,
					"Star Charter — reveal a creature card with power 3 or less and put it into your hand")),
		},
	})
}
