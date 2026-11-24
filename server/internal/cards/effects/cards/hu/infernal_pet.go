package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Infernal Pet — Creature — Imp {2}{B}, 2/2 (Kaldheim, #5505):
//
//	"Whenever you cast your second spell each turn, put a +1/+1 counter on
//	 this creature and it gains flying until end of turn."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "831b7dad-ab5a-4980-a52d-65c5633d474a",
		Name:         "Infernal Pet",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.YouCastYourSecondSpellEachTurn,
				"Infernal Pet — put a +1/+1 counter on this creature and it gains flying until end of turn",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1},
					effects.Instruction(effects.SourceGainsUntilEndOfTurn("flying", "Infernal Pet — flying until end of turn")))),
		},
	})
}
