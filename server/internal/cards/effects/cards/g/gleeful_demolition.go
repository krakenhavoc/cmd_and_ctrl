package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gleeful Demolition — Sorcery {R} (Phyrexia: All Will Be One):
//
//	"Destroy target artifact. If you controlled that artifact, create three
//	 1/1 red Phyrexian Goblin creature tokens."
//
// Whether the caster controlled the artifact is read as the spell resolves,
// just before the destruction, from the permanent as it is then (CR 608.2h).
// The tokens do not depend on the destruction itself: the text has no "if you
// do", so an indestructible artifact you control survives and the Goblins are
// made all the same. An illegal target means the spell does nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "66465bc1-4a34-4951-be53-57f90002976a",
		Name:         "Gleeful Demolition",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target artifact", effects.Artifact()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			c, ok := ctx.Game.LookupCardForEffect(id)
			if !ok {
				return nil
			}
			// "If you controlled that artifact": three Goblins, otherwise none.
			goblins := 0
			if c.Controller == ctx.Controller() {
				goblins = 3
			}
			if err := (effects.DestroyTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
			return effects.CreateToken{Controller: ctx.Controller(), Template: effects.TokenCard("1/1 red Phyrexian Goblin"), N: goblins}.Apply(ctx)
		},
	})
}
