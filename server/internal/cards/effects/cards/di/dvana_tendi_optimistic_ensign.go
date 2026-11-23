package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// D'Vana Tendi, Optimistic Ensign — Legendary Creature — Orion Scientist
// {1}{G/U}{G/U}, 2/4 (Star Trek, #4973):
//
//	"{T}: Untap another target permanent.
//	 {3}: Surveil 1."
//
// Ioreth of the Healing House's first ability, and a plain surveil sink.
// The untap is a CR 602 activation with a tap cost, so summoning
// sickness applies (CR 302.6); "another" keeps Tendi out of its own
// targets. The surveil needs no tap and can be activated any number of
// times.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6bb42689-b169-43cb-b787-f0e348063cd6",
		Name:         "D'Vana Tendi, Optimistic Ensign",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}: Untap another target permanent.",
				Cost:    effects.TapCost(),
				Targets: effects.Another(effects.TargetPermanent("another target permanent")),
				Effect:  effects.B22UntapEachLegalTarget,
			},
			{
				Label:   "{3}: Surveil 1.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.ManaCost("{3}"),
				Effect:  effects.Do(effects.Surveil{N: 1}),
			},
		},
	})
}
