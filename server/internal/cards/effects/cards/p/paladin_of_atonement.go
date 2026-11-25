package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Paladin of Atonement — Creature — Vampire Knight {1}{W}, 1/1:
//
//	"At the beginning of each upkeep, if you lost life last turn, put a +1/+1
//	 counter on this creature.
//	 When this creature dies, you gain life equal to its toughness."
//
// First Response's intervening "if" (CR 603.4, S82OldB083ILostLifeLastTurn):
// the previous turn's life loss, whoever's turn it was, checked as each upkeep
// begins and again as the ability resolves. The dies trigger reads the
// toughness the creature had as it last existed on the battlefield (CR 608.2h),
// counters included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f44dc71d-d08a-45ed-ac7c-f6d1ef811d96",
		Name:         "Paladin of Atonement",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.S82OldB083ILostLifeLastTurn,
				effects.On(game.EventBeginUpkeep, effects.S82OldB083ILostLifeLastTurnWhen,
					"Paladin of Atonement — put a +1/+1 counter on this creature",
					effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1}))),
			effects.WhenThisDies("Paladin of Atonement — you gain life equal to its toughness",
				effects.Do(effects.GainLifeEqualToSubject{Stat: effects.SubjectToughness})),
		},
	})
}
