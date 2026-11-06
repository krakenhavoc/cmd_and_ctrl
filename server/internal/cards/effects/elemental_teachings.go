package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Elemental Teachings — Instant — Lesson {4}{G}:
//
//	"Search your library for up to four land cards with different names
//	 and reveal them. An opponent chooses two of those cards. Put the
//	 chosen cards into your graveyard and the rest onto the battlefield
//	 tapped, then shuffle."
//
// Gifts Ungiven over land cards, with Emergent Ultimatum's untargeted
// "an opponent": the caster names which opponent chooses (ChoosePlayer),
// and that opponent picks two of the revealed lands. The pick is
// clamped to what was found (CR 608.2's "as much as possible"), so a
// search that found one or two cards sends them all to the graveyard
// (the ruling). The chosen go to the graveyard, the rest onto the
// battlefield tapped, and the library is shuffled last, as Gifts
// Ungiven explains: the found cards wait in the library, revealed, while
// the opponent answers, which changes nothing observable.
//
// No simplifications. Avatar: The Last Airbender (#4459).
func init() {
	Register(Spec{
		OracleID:     "7255705c-b788-4bb1-824a-9262f4a0864e",
		Name:         "Elemental Teachings",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Lands: 2},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return ctx.Game.SearchLibraryThenForEffect(game.SearchLibrarySpec{
				Player:   ctx.Controller(),
				Source:   ctx.Source(),
				Pred:     func(c game.Card) bool { return c.IsLand() },
				Dest:     game.ZoneLibrary,
				Limit:    4,
				Reveal:   true,
				Shuffle:  false,
				ToTop:    true,
				Reason:   "Elemental Teachings — search your library for up to four land cards with different names",
				Rule:     searchRuleDifferentNames,
				ThenTail: ctx.ItemTail(elementalTeachingsChooseTail),
			})
		},
	})
}

// elementalTeachingsChooseTail has the caster name the opponent who
// chooses. An on-disk key.
var elementalTeachingsChooseTail = CardTail("elemental-teachings/choose", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return ctx.Game.ShuffleLibraryForEffect(ctx.Controller())
	}
	return ChoosePlayer{
		Among:    Opponents,
		Question: "Elemental Teachings — choose an opponent to pick two of the revealed lands for the graveyard",
		ThenKey:  elementalTeachingsOpponent,
		Carry:    o.IDs,
	}.Apply(ctx)
})

// elementalTeachingsOpponent is the answer: that opponent picks two.
// Nobody to name puts every found land onto the battlefield, as no card
// was chosen. An on-disk key.
var elementalTeachingsOpponent = ChoosePlayerThen("option-pick/elemental-teachings-opponent", func(ctx *Context, chosen uuid.UUID, found []uuid.UUID) error {
	if chosen == uuid.Nil {
		return elementalTeachingsSettle(ctx, nil, found)
	}
	return RevealPick{
		Player:   chosen,
		Owner:    ctx.Controller(),
		Question: "Elemental Teachings — choose two of these lands; they go to their graveyard and the rest onto the battlefield tapped",
		Cards:    found,
		Min:      2,
		Max:      2,
		Keyed:    elementalTeachingsPicked,
	}.Apply(ctx)
})

var elementalTeachingsPicked = RegisterRevealPick("choose-cards/elemental-teachings-settle", func(ctx *Context, picked, left []uuid.UUID, _ []int) error {
	return elementalTeachingsSettle(ctx, picked, left)
})

// elementalTeachingsSettle puts the chosen lands into the graveyard and
// the rest onto the battlefield tapped, then shuffles.
func elementalTeachingsSettle(ctx *Context, picked, left []uuid.UUID) error {
	for _, id := range picked {
		if ctx.Game.FindCardZoneForEffect(id) == nil {
			continue
		}
		if err := ctx.Game.PutIntoGraveyardForEffect(id); err != nil {
			return err
		}
	}
	return ctx.Game.PutCardsFromLibraryOntoBattlefieldKeyedForEffect(liveCards(ctx.Game, left),
		game.LibraryEntryOptions{Tapped: true}, ctx.Tail(thenShuffle))
}
