package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bloodline Bidding — Sorcery {6}{B}{B} (Lorwyn Eclipsed, #4458):
//
//	"Convoke
//	 Choose a creature type. Return all creature cards of the chosen type
//	 from your graveyard to the battlefield."
//
// The caster names the type as the spell resolves (Kindred Judgment's
// ChooseCreatureTypeTail; nobody can respond to the choice). The cards are
// read after the choice and enter together under their owner's control
// (Raise the Past's ReturnFromGraveyardTogether, CR 603.6a), a changeling
// being of every type. Convoke is the shared tap cost (CR 702.51).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bdb9e933-f679-4930-b669-0d123db09d16",
		Name:         "Bloodline Bidding",
		Completeness: CompletenessFull,
		TapCost:      Convoke(),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			ChooseCreatureTypeTail(ctx.Game, item.Controller, item.SourceCardID, "Bloodline Bidding — choose a creature type",
				ctx.ItemTail(bloodlineBiddingTypeChosen))
			return nil
		},
	})
}

// bloodlineBiddingTypeChosen returns every creature card of the chosen type
// from the caster's graveyard to the battlefield. An on-disk key.
var bloodlineBiddingTypeChosen = creatureTypeChosen("then/bloodline-bidding-type-chosen", func(ctx *Context, t string, _ game.TailOutcome) error {
	return ReturnFromGraveyardTogether{
		Targets: graveyardCardIDs(ctx, ctx.Controller(), func(c game.Card) bool {
			return c.IsCreature() && c.HasSubtypeWord(t)
		}),
	}.Apply(ctx)
})
