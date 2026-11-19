package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thunder Drake — Creature — Elemental Drake {3}{U}, 2/3 (Ravnica: Clue Edition, #5929): "Flying. Whenever you cast your second spell each turn, put a +1/+1 counter on this creature."
//
// The second spell is the per-turn cast tally (Alphinaud Leveilleur).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a0956c3d-d382-486a-81f1-9aa5cf49b2ab",
		Name:            "Thunder Drake",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, YouCastYourSecondSpellEachTurn,
				"Thunder Drake — put a +1/+1 counter on this creature",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
