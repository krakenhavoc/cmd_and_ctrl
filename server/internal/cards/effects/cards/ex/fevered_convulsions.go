package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fevered Convulsions — Enchantment {B}{B} (Tempest, S82 OLD batch 760,
// #5628):
//
//	"{2}{B}{B}: Put a -1/-1 counter on target creature."
//
// Renew's counter-on-target body with a -1/-1 counter; a target that has
// left by resolution is illegal and nothing happens (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "83f35293-39e9-4c78-9c9f-25a1fc905676",
		Name:         "Fevered Convulsions",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{B}{B}: Put a -1/-1 counter on target creature.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.ManaCost("{2}{B}{B}"),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.RenewCountersOnTarget(game.CounterMinusOne, 1),
		}},
	})
}
