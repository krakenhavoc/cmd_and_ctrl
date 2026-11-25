package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reap the Past — Sorcery {X}{R}{G} (Modern Horizons):
//
//	"Return X cards at random from your graveyard to your hand. Exile Reap
//	 the Past."
//
// X is announced as the spell is cast (CR 107.3). The cards are chosen at
// random as the spell resolves, from the graveyard as it is then; if it
// holds fewer than X cards, they all return. The last sentence exiles the
// spell (a copy is not a card and ceases to exist), so Reap the Past never
// reaches the graveyard to be returned by a later copy of itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "27bc0219-3fc1-4c84-92cc-da2273a5fd16",
		Name:         "Reap the Past",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if x := ctx.X(); x > 0 {
				ids := effects.GraveyardIDs(ctx.Game, ctx.Controller(), nil)
				for _, id := range effects.RandomPick(ctx, ids, x) {
					if err := (effects.ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
						return err
					}
				}
			}
			return effects.ExileTarget{Target: ctx.Source()}.Apply(ctx)
		},
	})
}
