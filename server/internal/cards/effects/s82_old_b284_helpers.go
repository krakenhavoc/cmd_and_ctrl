package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b284_helpers.go — shared pieces of S82 older-set batch 284
// (Shadows over Innistrad: card draw and library, #5543).

// S82OldB284ForkInTheRoad is Fork in the Road's whole resolution: search for
// up to two basic land cards and reveal them, leave them on top (Intuition's
// shape, so the search reports what it found), let the searcher choose the
// one that goes to their hand, put the other into their graveyard, then
// shuffle.
func S82OldB284ForkInTheRoad(ctx *Context) error {
	return SearchLibrary{
		Player:    ctx.Controller(),
		Source:    ctx.Source(),
		Predicate: IsBasicLand,
		Dest:      game.ZoneLibrary,
		Limit:     2,
		Reveal:    true,
		Shuffle:   false,
		ToTop:     true,
		Reason:    "Fork in the Road — search your library for up to two basic land cards",
		ThenTail:  ctx.ItemTail(s82OldB284ForkChoose),
	}.Apply(ctx)
}

// s82OldB284ForkChoose is the search's continuation: nothing found only
// shuffles, and otherwise the searcher picks the card for their hand. An
// on-disk key.
var s82OldB284ForkChoose = CardTail("fork-in-the-road/choose", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return ctx.Game.ShuffleLibraryForEffect(ctx.Controller())
	}
	return RevealPick{
		Player:   ctx.Controller(),
		Owner:    ctx.Controller(),
		Question: "Fork in the Road — choose one of these cards; it goes to your hand and the other to your graveyard",
		Cards:    o.IDs,
		Min:      1,
		Max:      1,
		Keyed:    s82OldB284ForkSettle,
	}.Apply(ctx)
})

// s82OldB284ForkSettle moves the chosen card to the hand, the other to the
// graveyard, and shuffles. An on-disk key.
var s82OldB284ForkSettle = RegisterRevealPick("choose-cards/fork-in-the-road-settle", func(ctx *Context, picked, left []uuid.UUID, _ []int) error {
	return revealPickSettle(ctx, picked, left, true)
})
