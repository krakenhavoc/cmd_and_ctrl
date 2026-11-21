package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thunder Drake — Creature — Elemental Drake {3}{U}, 2/3 (Ravnica: Clue Edition, #5929): "Flying. Whenever you cast your second spell each turn, put a +1/+1 counter on this creature."
//
// The second spell is the per-turn cast tally (Alphinaud Leveilleur).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a0956c3d-d382-486a-81f1-9aa5cf49b2ab",
		Name:            "Thunder Drake",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.YouCastYourSecondSpellEachTurn,
				"Thunder Drake — put a +1/+1 counter on this creature",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
