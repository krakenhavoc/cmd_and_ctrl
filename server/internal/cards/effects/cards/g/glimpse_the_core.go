package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glimpse the Core — Sorcery {1}{G} (The Lost Caverns of Ixalan, #4470):
//
//	"Choose one —
//	 • Search your library for a basic Forest card, put that card onto
//	   the battlefield tapped, then shuffle.
//	 • Return target Cave card from your graveyard to the battlefield
//	   tapped."
//
// The mode is chosen as the spell is cast (CR 700.2a), and only the
// second bullet has a target, so a graveyard with no Cave card leaves the
// first bullet as the only choice. The search is for a card with both the
// basic supertype and the Forest subtype (CR 701.23a), and may find
// nothing. The returned Cave enters tapped; one that left the graveyard
// in response is no longer a legal target and the spell does nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fa44fa48-d270-4f0f-b672-d63e1c1942ce",
		Name:         "Glimpse the Core",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Search your library for a basic Forest card, put that card onto the battlefield tapped, then shuffle.", nil,
				func(item *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.SearchLibrary{
						Player:        item.Controller,
						Predicate:     effects.IsBasicLandWithSubtype("forest"),
						Dest:          game.ZoneBattlefield,
						Limit:         1,
						TappedOnEntry: true,
						Shuffle:       true,
						Reason:        "Glimpse the Core — a basic Forest card, onto the battlefield tapped",
					}.Apply(ctx)
				}), game.Purpose{Lands: 1}),
			effects.ModeDoing("Return target Cave card from your graveyard to the battlefield tapped.",
				effects.TargetCardInGraveyard("target Cave card from your graveyard", effects.HasSubtype("Cave"), effects.YouOwn()),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					if t, ok := effects.ModeTarget(ctx, occ); ok && t.Kind == game.TargetCard {
						return effects.ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield, Tapped: true}.Apply(ctx)
					}
					return nil
				}),
		),
	})
}
