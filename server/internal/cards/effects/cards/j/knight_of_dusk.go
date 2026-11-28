package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Knight of Dusk — Creature — Human Knight {1}{B}{B}, 2/2 (Tempest, S82 OLD,
// #5628):
//
//	"{B}{B}: Destroy target creature blocking this creature."
//
// "Blocking this creature" is BlockingSource, judged at announce and again at
// resolution, so a blocker removed from combat in response is an illegal target
// (CR 608.2b) and a Knight that has left the battlefield has no blockers.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cd1ca3c8-1f03-4c03-a93a-0419ac205c18",
		Name:         "Knight of Dusk",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}{B}: Destroy target creature blocking this creature.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.ManaCost("{B}{B}"),
			Targets: effects.BlockingSource(effects.TargetCreature("target creature blocking this creature")),
			Effect:  effects.DestroyFirstLegalTarget,
		}},
	})
}
