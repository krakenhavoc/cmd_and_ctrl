package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sanguine Sacrament — Instant {X}{W}{W} (Ixalan, S82 older sets):
//
//	"You gain twice X life. Put Sanguine Sacrament on the bottom of its
//	 owner's library."
//
// The spell puts itself on the bottom as it resolves (Spell Crumple's
// PutIntoLibrary), instead of going to the graveyard; a copy does not (it is
// no card, CR 707.10a). X may be 0: the spell then gains nothing and still
// goes to the library.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "851be993-35d7-4566-96a0-3fe84d17979e",
		Name:         "Sanguine Sacrament",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := (effects.GainLife{Amount: 2 * ctx.X()}).Apply(ctx); err != nil {
				return err
			}
			if item.IsCopy {
				return nil
			}
			return effects.PutIntoLibrary{Card: ctx.Source(), ToBottom: true}.Apply(ctx)
		},
	})
}
