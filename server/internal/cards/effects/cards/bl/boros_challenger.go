package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Boros Challenger — Creature — Human Soldier {R}{W}, 2/3 (Guilds of
// Ravnica):
//
//	"Mentor (Whenever this creature attacks, put a +1/+1 counter on target
//	 attacking creature with lesser power.)
//	 {2}{R}{W}: This creature gets +1/+1 until end of turn."
//
// The shared mentor trigger (CR 702.134a) and Aetherwind Basker's self pump as
// an activated row.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bc6b40b1-0d17-4dd2-bec9-f1653611dba3",
		Name:         "Boros Challenger",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.Mentor("Boros Challenger — mentor")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R}{W}: This creature gets +1/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: 1}},
			Cost:    effects.ManaCost("{2}{R}{W}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, 1, "Boros Challenger — +1/+1 until end of turn"),
		}},
	})
}
