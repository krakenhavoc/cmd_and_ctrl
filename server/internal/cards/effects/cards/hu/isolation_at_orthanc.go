package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Isolation at Orthanc — Instant {3}{U} (The Lord of the Rings: Tales of
// Middle-earth, #5948):
//
//	"Put target creature into its owner's library second from the top."
//
// Oust's move without the life gain: the card goes into its owner's library,
// not the caster's, second from the top (a library of one card takes it on the
// bottom, which is also second from the top). A token ceases to exist.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6dc6536f-455a-4d7b-beed-a78ec63c0030",
		Name:         "Isolation at Orthanc",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			return effects.PutIntoLibrary{Card: t.ID, Depth: 2}.Apply(ctx)
		},
	})
}
