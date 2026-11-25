package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Old-Growth Dryads — Creature — Dryad {G}, 3/3 (Ixalan, S82 older sets):
//
//	"When this creature enters, each opponent may search their library for a
//	 basic land card, put it onto the battlefield tapped, then shuffle."
//
// Rootweaver Druid's per-opponent search, in turn order, with one card, no
// FirstEnters (the land enters under the searcher's control) and every
// opponent shuffling.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aa125d1c-ab8f-48ff-a6eb-d8fc2154322c",
		Name:         "Old-Growth Dryads",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Old-Growth Dryads — each opponent may search their library for a basic land card and put it onto the battlefield tapped",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, opp := range effects.OpponentsInTurnOrder(g, item.Controller) {
						if err := (effects.SearchLibrary{
							Player:        opp,
							Predicate:     effects.IsBasicLand,
							Dest:          game.ZoneBattlefield,
							Limit:         1,
							Reveal:        true,
							Optional:      true,
							Shuffle:       true,
							TappedOnEntry: true,
							Reason:        "Old-Growth Dryads — a basic land card onto the battlefield tapped",
						}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
		},
	})
}
