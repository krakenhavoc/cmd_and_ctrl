package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Plunder — Sorcery {4}{R} (Time Spiral, #5596):
//
//	"Destroy target artifact or land.
//	 Suspend 4—{1}{R} (Rather than cast this card from your hand, you
//	 may pay {1}{R} and exile it with four time counters on it. At the
//	 beginning of your upkeep, remove a time counter. When the last is
//	 removed, you may cast it without paying its mana cost.)"
//
// Pillage's target on a suspend card, without Pillage's "can't be
// regenerated": a regeneration shield saves the target here.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8a515f67-f868-494b-8bd7-de63fe14477b",
		Name:         "Plunder",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target artifact or land", effects.Or(effects.Artifact(), effects.Land())),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		SpecialActions: []game.SpecialAction{
			effects.Suspend(4, "{1}{R}"),
		},
		OnResolve: effects.DestroyTheTargetPermanent,
	})
}
