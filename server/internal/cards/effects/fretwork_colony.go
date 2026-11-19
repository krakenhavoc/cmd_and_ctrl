package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fretwork Colony — Creature — Insect {1}{B}, 1/1:
//
//	"This creature can't block.
//	 At the beginning of your upkeep, put a +1/+1 counter on this creature
//	 and you lose 1 life."
//
// "Can't block" is a restriction printed on the creature itself (Aesthir
// Glider). The upkeep trigger is one ability: the counter, then the life loss.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "95f5ddb6-e3e5-4b35-88ac-eae192783eda",
		Name:         "Fretwork Colony",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{RestrictSelf(game.CantBlock)},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Fretwork Colony — put a +1/+1 counter on it and you lose 1 life", Do(
				CounterOnThis{Kind: game.CounterPlusOne, N: 1},
				GainLife{Amount: -1},
			)),
		},
	})
}
