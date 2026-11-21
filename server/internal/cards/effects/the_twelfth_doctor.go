package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Twelfth Doctor — Legendary Creature — Time Lord Doctor {3}{U}{R}, 4/4
// (Doctor Who, #5948):
//
//	"The first spell you cast from anywhere other than your hand each turn
//	 has demonstrate. (When you cast that spell, you may copy it. If you do,
//	 choose an opponent to also copy it. A copy of a permanent spell becomes
//	 a token.)
//	 Whenever you copy a spell, put a +1/+1 counter on The Twelfth Doctor."
//
// Granting demonstrate to a spell as it is cast is the same as a trigger of
// the Doctor's that fires on that cast (CR 702.144a's cast trigger is only
// ever on the stack object; nothing else reads it): when the first spell you
// cast from a zone other than your hand this turn is cast (read off the cast
// tally, which is bumped before the cast event), you may copy it, and if you
// do an opponent you choose copies it too (Demonstrate's own body, applied
// to that spell). The second ability counts only copies YOU make (Parnesse's
// reading of EventSpellCopied), so the opponent's demonstrate copy and a
// copy an opponent makes of their own spells do not grow the Doctor, and
// each copy of a spell is one trigger.
//
// No simplification.
func init() {
	demonstrate := Optional(On(game.EventCast, gapB89FirstCastNotFromHandThisTurn,
		"The Twelfth Doctor — demonstrate: you may copy that spell; if you do, an opponent of your choice also copies it",
		gapB89DemonstrateTheSpellCast),
		"The Twelfth Doctor — copy that spell? If you do, choose an opponent to also copy it.")
	Register(Spec{
		OracleID:     "9ecf19ba-754e-4bb3-b526-fe9df87a0f08",
		Name:         "The Twelfth Doctor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			demonstrate,
			On(game.EventSpellCopied, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventSpellCopied && ev.Actor == source.Controller
			}, "The Twelfth Doctor — put a +1/+1 counter on it",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
