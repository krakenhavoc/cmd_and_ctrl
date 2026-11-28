package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fumarole — Sorcery {3}{B}{R} (Ice Age, S82 OLD, #5636):
//
//	"As an additional cost to cast this spell, pay 3 life.
//	 Destroy target creature and target land."
//
// Spiteful Blow's two clauses (one target each, re-checked on their own at resolution, CR 608.2b), with
// the life paid as the additional cost (CR 118.4: it can't be cast with fewer than 3 life).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "a9e33138-e1c0-4b69-b868-1ce8493c073d",
		Name:           "Fumarole",
		Completeness:   effects.CompletenessFull,
		Purpose:        game.Purpose{Answers: game.AnswerRemove},
		AdditionalCost: effects.PayLifeCost(3),
		Targets: effects.Clauses(
			effects.TargetCreature("target creature"),
			effects.TargetPermanent("target land", effects.Land()),
		),
		OnResolve: effects.DestroyEachLegalTarget,
	})
}
