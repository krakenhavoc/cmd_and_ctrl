package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "fa44fa48-d270-4f0f-b672-d63e1c1942ce",
		Name:         "Glimpse the Core",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeWithPurpose(ModeDoing("Search your library for a basic Forest card, put that card onto the battlefield tapped, then shuffle.", nil,
				func(item *game.StackItem, ctx *Context, _ int) error {
					return SearchLibrary{
						Player:        item.Controller,
						Predicate:     IsBasicLandWithSubtype("forest"),
						Dest:          game.ZoneBattlefield,
						Limit:         1,
						TappedOnEntry: true,
						Shuffle:       true,
						Reason:        "Glimpse the Core — a basic Forest card, onto the battlefield tapped",
					}.Apply(ctx)
				}), game.Purpose{Lands: 1}),
			ModeDoing("Return target Cave card from your graveyard to the battlefield tapped.",
				TargetCardInGraveyard("target Cave card from your graveyard", HasSubtype("Cave"), YouOwn()),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					if t, ok := ModeTarget(ctx, occ); ok && t.Kind == game.TargetCard {
						return ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield, Tapped: true}.Apply(ctx)
					}
					return nil
				}),
		),
	})
}
