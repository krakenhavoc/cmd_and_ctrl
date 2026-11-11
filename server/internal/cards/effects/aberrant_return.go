package effects

import (
	"errors"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aberrant Return — Sorcery {4}{B}{B} (Lorwyn Eclipsed Commander, #5368):
//
//	"Put one, two, or three target creature cards from graveyards onto the
//	 battlefield under your control. Each of them enters with an
//	 additional -1/-1 counter on it."
//
// One clause of one to three creature cards in any graveyards. Each card
// still there as the spell resolves comes back under the caster's control
// with the counter as an "enters with" (CR 614.1c), so it is a creature with
// the counter already on it when any enters trigger or state check sees it.
// A card that left its graveyard in response is skipped (CR 608.2b, 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3a92b235-196b-4f46-9d20-06f4d3653d36",
		Name:         "Aberrant Return",
		Completeness: CompletenessFull,
		Targets:      TargetCardInGraveyard("one, two, or three target creature cards from graveyards", Creature()).WithCount(1, 3),
		Purpose:      ForTargets(Reanimates(0)),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, id := range legalTargetCardIDs(ctx) {
				_, err := ctx.Game.ReturnFromGraveyardWithCountersForEffect(id, ctx.Controller(), false,
					map[string]int{game.CounterMinusOne: 1})
				if err != nil && !errors.Is(err, game.ErrCardNotFound) {
					return err
				}
			}
			return nil
		},
	})
}
