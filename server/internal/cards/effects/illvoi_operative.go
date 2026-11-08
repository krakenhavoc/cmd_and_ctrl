package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Illvoi Operative — Creature — Jellyfish Rogue {1}{U}, 2/1 (Edge of
// Eternities, #4461):
//
//	"Whenever you cast your second spell each turn, put a +1/+1 counter
//	 on this creature."
//
// Sunstar Lightsmith's trigger (YouCastYourSecondSpellEachTurn) without
// the draw: the second spell you cast each turn, and not the third.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2e0b225a-d4b4-4f53-8fea-8c6c37102920",
		Name:         "Illvoi Operative",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, YouCastYourSecondSpellEachTurn,
				"Illvoi Operative — put a +1/+1 counter on this creature",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
