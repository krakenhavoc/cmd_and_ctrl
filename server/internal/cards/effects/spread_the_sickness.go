package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spread the Sickness — Sorcery {4}{B} (Mirrodin Besieged,
// #5576):
//
//	"Destroy target creature, then proliferate. (Choose any number of
//	 permanents and/or players, then give each another counter of each
//	 kind already there.)"
//
// Murder's destroy, then the player's proliferate choice (Atomize's
// shape): an indestructible target survives and the proliferate still
// happens. With its only target gone the spell does not resolve at all
// (CR 608.2b), so there is no proliferate either.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "142af8f1-6d6b-4043-8cc4-6632481972c1",
		Name:         "Spread the Sickness",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := destroyTheTargetPermanent(item, ctx); err != nil {
				return err
			}
			return Proliferate{}.Apply(ctx)
		},
	})
}
