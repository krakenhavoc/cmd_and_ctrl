package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Joraga Auxiliary — Creature — Elf Soldier Ally {1}{G}{W}, 2/3 (Oath of the
// Gatewatch, #5545):
//
//	"{4}{G}{W}: Support 2. (Put a +1/+1 counter on each of up to two other
//	 target creatures.)"
//
// A permanent's support targets OTHER creatures (CR 701.41a): up to two
// creatures other than the Auxiliary, a counter on each still-legal one as the
// ability resolves (CR 608.2b). It may be activated with no target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b74082da-435c-4e45-9424-1188ad7ef5b3",
		Name:         "Joraga Auxiliary",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{G}{W}: Support 2.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{4}{G}{W}"),
			Targets: effects.SupportTargets(2, true),
			Effect:  effects.SupportEffect,
		}},
	})
}
