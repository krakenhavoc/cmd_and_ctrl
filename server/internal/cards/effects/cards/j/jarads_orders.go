package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jarad's Orders — Sorcery {2}{B}{G} (Return to Ravnica, S82 OLD):
//
//	"Search your library for up to two creature cards and reveal them. Put
//	 one into your hand and the other into your graveyard. Then shuffle."
//
// Intuition's search and settle with the caster doing the choosing. The
// search is for up to two creature cards (a quality, so it may find fewer);
// the cards wait in the library, revealed, while the caster picks the one
// that goes to their hand, and the other goes to their graveyard, then the
// library is shuffled. One creature found goes to the hand (the pick is
// clamped to what was found); none found only shuffles.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "93036f4e-f553-4cc7-a747-6cd087b0994c",
		Name:         "Jarad's Orders",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 2},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return ctx.Game.SearchLibraryThenForEffect(game.SearchLibrarySpec{
				Player:   ctx.Controller(),
				Source:   ctx.Source(),
				Pred:     func(c game.Card) bool { return c.IsCreature() },
				Dest:     game.ZoneLibrary,
				Limit:    2,
				Reveal:   true,
				Shuffle:  false,
				ToTop:    true,
				Reason:   "Jarad's Orders — search your library for up to two creature cards",
				ThenTail: ctx.ItemTail(jaradsOrdersChoose),
			})
		},
	})
}

// jaradsOrdersChoose is the search's continuation: the caster picks the card
// that goes to their hand. An on-disk key.
var jaradsOrdersChoose = effects.CardTail("jarads-orders/choose", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return ctx.Game.ShuffleLibraryForEffect(ctx.Controller())
	}
	return effects.RevealPick{
		Player:   ctx.Controller(),
		Owner:    ctx.Controller(),
		Question: "Jarad's Orders — choose one of these creature cards to put into your hand; the other goes to your graveyard",
		Cards:    o.IDs,
		Min:      1,
		Max:      1,
		Keyed:    jaradsOrdersSettle,
	}.Apply(ctx)
})

// jaradsOrdersSettle puts the chosen card into the hand and the other into
// the graveyard, then shuffles. A card that is no longer where the search
// left it is skipped. An on-disk key.
var jaradsOrdersSettle = effects.RegisterRevealPick("choose-cards/jarads-orders-settle", func(ctx *effects.Context, picked, left []uuid.UUID, _ []int) error {
	for _, id := range left {
		if ctx.Game.FindCardZoneForEffect(id) == nil {
			continue
		}
		if err := ctx.Game.PutIntoGraveyardForEffect(id); err != nil {
			return err
		}
	}
	return ctx.Game.BounceCardsToHandKeyedForEffect(effects.LiveCards(ctx.Game, picked), ctx.Tail(effects.ThenShuffle))
})
