package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mobile Garrison — Artifact — Vehicle {3}, 3/4 (Aether Revolt, S82 older
// sets, #5536):
//
//	"Whenever this Vehicle attacks, untap another target artifact or
//	 creature you control.
//	 Crew 2"
//
// "Another" is object identity (the Vehicle itself is an artifact creature
// while it attacks and is never offered). The target is chosen as the
// trigger goes on the stack and re-checked as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "14ef5aa4-5f61-4fbd-9a8d-35920c40ec25",
		Name:         "Mobile Garrison",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverThisAttacks("Mobile Garrison — untap another target artifact or creature you control",
				effects.UntapFirstLegalTarget),
				effects.Another(effects.TargetPermanent("another target artifact or creature you control",
					effects.Or(effects.Artifact(), effects.Creature()), effects.YouControl()))),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    effects.CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Mobile Garrison"),
		}},
	})
}
