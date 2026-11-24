package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bloodsky Berserker — Creature — Human Berserker {1}{B}, 1/1 (Kaldheim, #5505):
//
//	"Whenever you cast your second spell each turn, put two +1/+1 counters
//	 on this creature. It gains menace until end of turn."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ed04f170-a4d9-4151-b4ae-1200d523dc6b",
		Name:         "Bloodsky Berserker",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.YouCastYourSecondSpellEachTurn,
				"Bloodsky Berserker — put two +1/+1 counters on this creature; it gains menace until end of turn",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 2},
					effects.Instruction(effects.SourceGainsUntilEndOfTurn("menace", "Bloodsky Berserker — menace until end of turn")))),
		},
	})
}
