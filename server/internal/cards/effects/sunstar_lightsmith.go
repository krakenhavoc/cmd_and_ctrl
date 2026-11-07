package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sunstar Lightsmith — Creature — Human Artificer {3}{W}, 3/3 (Edge of
// Eternities, #4461):
//
//	"Whenever you cast your second spell each turn, put a +1/+1 counter
//	 on this creature and draw a card."
//
// Devoted Duelist's second-spell watch (YouCastYourSecondSpellEachTurn,
// which counts the spells cast this turn, so a spell cast before the
// Lightsmith arrived counts toward the two). The counter goes on this
// creature only while it is still the same object (CounterOnThis); the
// draw happens either way.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "132de6c2-4ca2-41da-9c78-caa59838b641",
		Name:         "Sunstar Lightsmith",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, YouCastYourSecondSpellEachTurn,
				"Sunstar Lightsmith — put a +1/+1 counter on this creature and draw a card",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1}, DrawCards{N: 1})),
		},
	})
}
