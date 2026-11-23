package cardsu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ultimate Nullification — {4}{W} Sorcery (Marvel Super Heroes Commander,
// #4455):
//
//	"As an additional cost to cast this spell, sacrifice a legendary
//	 creature.
//	 Exile all creatures and graveyards. Put Ultimate Nullification on the
//	 bottom of its owner's library."
//
// Final Judgment's creature exile, then every seat's graveyard through
// Farewell's helper, then the spell tucks itself on the bottom (Approach of
// the Second Sun's route with TuckOptions.ToBottom): the spell is still on the
// stack while it does, so it never reaches the graveyard it just emptied.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "2fe0ebf5-52ed-4e92-9b93-81e9ae564439",
		Name:           "Ultimate Nullification",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.SacrificeCost("a legendary creature", effects.Legendary(), effects.Creature()),
		Purpose:        game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepExile}},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := (effects.ExileAllMatching{Match: effects.Creature()}).Apply(ctx); err != nil {
				return err
			}
			for _, p := range ctx.Game.Seats {
				if p == nil {
					continue
				}
				if err := effects.ExileGraveyardForEffect(ctx.Game, item, p.ID); err != nil {
					return err
				}
			}
			return ctx.Game.TuckCardsToLibraryKeyedForEffect([]uuid.UUID{ctx.Source()}, game.TuckOptions{ToBottom: true}, game.KeyedTail{})
		},
	})
}
