package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Poised Practitioner — Creature — Human Monk {2}{W}, 2/3 (Tarkir:
// Dragonstorm, #4463):
//
//	"Flurry — Whenever you cast your second spell each turn, put a +1/+1
//	 counter on this creature. Scry 1. (Look at the top card of your
//	 library. You may put that card on the bottom.)"
//
// Flurry is an ability word (CR 207.2c), not a keyword: the trigger is
// YouCastYourSecondSpellEachTurn, Cori-Steel Cutter's. Spells cast before
// the Practitioner was on the battlefield count (the printed ruling), so
// a Practitioner that was itself the first spell of the turn triggers on
// the next one. The counter and the scry (CR 701.22a) are one Do, so a
// counter placement that asks a question holds the scry until it is
// answered.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e08debdc-2db5-4562-8a02-4158bfc0db88",
		Name:         "Poised Practitioner",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, YouCastYourSecondSpellEachTurn,
				"Poised Practitioner — flurry: put a +1/+1 counter on this creature, then scry 1",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1}, Scry{N: 1})),
		},
	})
}
