package cardsfl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Footbottom Feast — Instant {2}{B} (Commander 2011, #5574):
//
//	"Put any number of target creature cards from your graveyard on top of
//	 your library.
//	 Draw a card."
//
// "Any number" includes none, so the spell is castable with no targets and
// still draws. The cards go on top together (CR 401.4) and the draw is the
// move's continuation, so it happens after they land; with every chosen
// target gone by resolution the spell does nothing at all (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0a35df9d-80a0-4b6f-91d5-dad980af47e1",
		Name:         "Footbottom Feast",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCardInGraveyard("any number of target creature cards from your graveyard", effects.Creature(), effects.YouOwn()).WithCount(0, 0),
		Purpose:      game.Purpose{Draws: 1},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			var ids []uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					ids = append(ids, t.ID)
				}
			}
			if len(ids) == 0 {
				return effects.DrawCards{N: 1}.Apply(ctx)
			}
			return ctx.Game.TuckCardsToLibraryKeyedForEffect(ids, game.TuckOptions{}, ctx.TailN(effects.ThenDrawCards, 1))
		},
	})
}
