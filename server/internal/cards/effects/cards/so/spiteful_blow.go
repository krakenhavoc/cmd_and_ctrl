package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spiteful Blow — Sorcery {4}{B}{B} (Journey into Nyx, S82 OLD, #5558):
//
//	"Destroy target creature and target land."
//
// Two target clauses, each judged on its own at resolution (CR 608.2b): if one
// is gone the other is still destroyed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "25a20f94-1703-4bcd-862d-f2db7014ba90",
		Name:         "Spiteful Blow",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: effects.Clauses(
			effects.TargetCreature("target creature"),
			effects.TargetPermanent("target land", effects.Land()),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for slot := 0; slot < 2; slot++ {
				if t, ok := ctx.ClauseTarget(slot); ok && t.Kind == game.TargetCard && ctx.IsTargetLegal(t) {
					if err := (effects.DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
						return err
					}
				}
			}
			return nil
		},
	})
}
