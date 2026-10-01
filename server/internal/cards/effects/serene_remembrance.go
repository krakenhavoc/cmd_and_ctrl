package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Serene Remembrance — Sorcery {G}:
//
//	"Shuffle Serene Remembrance and up to three target cards from a
//	 single graveyard into their owners' libraries."
//
// #1807, ADR 0106 §5. The spell and its still-legal targets go into
// their owners' libraries in one move, and then each of those
// libraries is shuffled. CR 701.24c: a library an effect says to
// shuffle objects into is shuffled even if none of them arrived, so
// the shuffle reads the OWNERS, not what landed.
//
// The 2013-01-24 rulings: with no target chosen, only Serene
// Remembrance goes into its owner's library; with targets chosen and
// every one illegal on resolution, the spell does not resolve at all
// (CR 608.2b) and goes to the graveyard, which the engine's fizzle
// already does before OnResolve runs.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f980c8b4-4cd7-42d2-82d5-96e3275b2647",
		Name:         "Serene Remembrance",
		Completeness: CompletenessFull,
		Targets:      upToNCardsFromASingleGraveyard(3),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			ids := []uuid.UUID{ctx.Source()}
			owners := []uuid.UUID{item.Owner}
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				card, ok := ctx.Game.LookupCardForEffect(t.ID)
				if !ok {
					continue
				}
				ids = append(ids, t.ID)
				if card.Owner != item.Owner && (len(owners) < 2 || owners[1] != card.Owner) {
					owners = append(owners, card.Owner)
				}
			}
			return ctx.Game.TuckCardsToLibraryThenForEffect(ids, game.TuckOptions{}, func(g *game.Game, _ []uuid.UUID) error {
				for _, owner := range owners {
					if err := g.ShuffleLibraryForEffect(owner); err != nil {
						return err
					}
				}
				return nil
			})
		},
	})
}
