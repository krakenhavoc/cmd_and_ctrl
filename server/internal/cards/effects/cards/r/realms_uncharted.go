package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Realms Uncharted — Instant {2}{G} (Secret Lair Drop):
//
//	"Search your library for up to four land cards with different names and
//	 reveal them. An opponent chooses two of those cards. Put the chosen cards
//	 into your graveyard and the rest into your hand. Then shuffle."
//
// Gifts Ungiven over land cards (Elemental Teachings' body with the rest going
// to the hand): the caster names which opponent chooses (ChoosePlayer) and that
// opponent picks two of the revealed lands. The pick is clamped to what was
// found (CR 608.2's "as much as possible"), so a search that found one or two
// cards sends them all to the graveyard. The library is shuffled last, the
// found cards waiting in it, revealed, while the opponent answers.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e21c8fc6-d4ef-42b6-b11e-d9c931da1387",
		Name:         "Realms Uncharted",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return ctx.Game.SearchLibraryThenForEffect(game.SearchLibrarySpec{
				Player:   ctx.Controller(),
				Source:   ctx.Source(),
				Pred:     func(c game.Card) bool { return c.IsLand() },
				Dest:     game.ZoneLibrary,
				Limit:    4,
				Reveal:   true,
				Shuffle:  false,
				ToTop:    true,
				Reason:   "Realms Uncharted — search your library for up to four land cards with different names",
				Rule:     effects.SearchRuleDifferentNames,
				ThenTail: ctx.ItemTail(realmsUnchartedChooseTail),
			})
		},
	})
}

// realmsUnchartedChooseTail has the caster name the opponent who chooses. An
// on-disk key: never renamed, never reused.
var realmsUnchartedChooseTail = effects.CardTail("realms-uncharted/choose", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return ctx.Game.ShuffleLibraryForEffect(ctx.Controller())
	}
	return effects.ChoosePlayer{
		Among:    effects.Opponents,
		Question: "Realms Uncharted — choose an opponent to pick two of the revealed lands for the graveyard",
		ThenKey:  realmsUnchartedOpponent,
		Carry:    o.IDs,
	}.Apply(ctx)
})

// realmsUnchartedOpponent is the answer: that opponent picks two. Nobody to
// name sends every found land to the hand, as no card was chosen. An on-disk
// key: never renamed, never reused.
var realmsUnchartedOpponent = effects.ChoosePlayerThen("option-pick/realms-uncharted-opponent", func(ctx *effects.Context, chosen uuid.UUID, found []uuid.UUID) error {
	if chosen == uuid.Nil {
		return realmsUnchartedSettle(ctx, nil, found)
	}
	return effects.RevealPick{
		Player:   chosen,
		Owner:    ctx.Controller(),
		Question: "Realms Uncharted — choose two of these lands; they go to their graveyard and the rest to their hand",
		Cards:    found,
		Min:      2,
		Max:      2,
		Keyed:    realmsUnchartedPicked,
	}.Apply(ctx)
})

var realmsUnchartedPicked = effects.RegisterRevealPick("choose-cards/realms-uncharted-settle", func(ctx *effects.Context, picked, left []uuid.UUID, _ []int) error {
	return realmsUnchartedSettle(ctx, picked, left)
})

// realmsUnchartedSettle puts the chosen lands into the graveyard and the rest
// into the hand, then shuffles.
func realmsUnchartedSettle(ctx *effects.Context, picked, left []uuid.UUID) error {
	for _, id := range picked {
		if ctx.Game.FindCardZoneForEffect(id) == nil {
			continue
		}
		if err := ctx.Game.PutIntoGraveyardForEffect(id); err != nil {
			return err
		}
	}
	return ctx.Game.TakeFromLibraryToHandKeyedForEffect(ctx.Controller(), effects.LiveCards(ctx.Game, left), ctx.Tail(effects.ThenShuffle))
}
