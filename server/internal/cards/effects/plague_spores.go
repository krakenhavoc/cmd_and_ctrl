package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Plague Spores — Sorcery {4}{B}{R} (Invasion, #5616):
//
//	"Destroy target nonblack creature and target land. They can't be
//	 regenerated."
//
// Two target clauses, Decimate's shape: each still-legal target is
// destroyed at the same time, with regeneration ignored (CR 701.19c).
// One that has become illegal is left alone and the other is still
// destroyed; with both gone the spell does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f79f5cb7-d238-41ab-85f0-eb21cc6bf561",
		Name:         "Plague Spores",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: TargetCreature("target nonblack creature", NonBlack()).Then(
			TargetPermanent("target land", Land()),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			var ids []uuid.UUID
			for _, t := range item.Targets {
				if t.Kind == game.TargetCard && ctx.IsTargetLegal(t) {
					ids = append(ids, t.ID)
				}
			}
			if len(ids) > 0 {
				ctx.Game.DestroyPermanentsForEffect(ids, game.DestroyOptions{CantBeRegenerated: true})
			}
			return nil
		},
	})
}
