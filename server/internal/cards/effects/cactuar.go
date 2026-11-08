package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cactuar — Creature — Plant {G}, 3/3 (Final Fantasy, #4462):
//
//	"Trample
//	 At the beginning of your end step, if this creature didn't enter the
//	 battlefield this turn, return it to its owner's hand."
//
// An intervening "if" (CR 603.4): the trigger is put on the stack only
// when the end step begins on its controller's turn and Cactuar did not
// enter this turn, and the condition is read again as it resolves.
// "Entered this turn" is the turn tally's entry count for this object
// (Game.EnteredThisTurn), not summoning sickness, so a Cactuar that
// entered on an opponent's turn goes back at the end of yours. The
// return reads the source as the object that triggered (CR 400.7): one
// that left and came back in response stays where it is.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3e61d7c4-263f-4077-bdbc-d79020f2a479",
		Name:            "Cactuar",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginEndStep, AllOf(ByYou, cactuarDidNotEnterThisTurn),
				"Cactuar — return it to its owner's hand", cactuarReturn),
		},
	})
}

// cactuarDidNotEnterThisTurn is the condition checked as the end step
// begins.
func cactuarDidNotEnterThisTurn(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return source != nil && !g.EnteredThisTurn(source.InstanceID)
}

// cactuarReturn checks the condition again as the trigger resolves, then
// returns Cactuar to its owner's hand.
func cactuarReturn(g *game.Game, item *game.StackItem) error {
	if g.EnteredThisTurn(item.SourceCardID) {
		return nil
	}
	return returnThisPermanentToOwnersHand(g, item)
}
