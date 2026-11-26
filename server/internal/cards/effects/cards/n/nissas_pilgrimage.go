package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nissa's Pilgrimage — Sorcery {2}{G} (Commander 2017, S82 older sets):
//
//	"Search your library for up to two basic Forest cards, reveal those cards,
//	 and put one onto the battlefield tapped and the rest into your hand. Then
//	 shuffle.
//	 Spell mastery — If there are two or more instant and/or sorcery cards in
//	 your graveyard, search your library for up to three basic Forest cards
//	 instead of two."
//
// Cultivate's two-step search (TwoLandSearch): one basic Forest onto the
// battlefield tapped, then the rest into the hand (one, or two with spell
// mastery, which is read as the spell resolves, CR 608.2c, with this spell
// still on the stack and so not itself in the graveyard) and the shuffle.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e83236f8-1c45-4d31-8937-1c76db549955",
		Name:         "Nissa's Pilgrimage",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Lands: 1, Tutors: 1},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			hand := 1
			if effects.SpellMastery(ctx.Game, item.Controller) {
				hand = 2
			}
			forest := effects.IsBasicLandOfAnySubtype("Forest")
			return effects.SearchLibrary{
				Player:        item.Controller,
				Source:        ctx.Source(),
				Predicate:     forest,
				Dest:          game.ZoneBattlefield,
				Limit:         1,
				Reveal:        true,
				TappedOnEntry: true,
				Reason:        "Nissa's Pilgrimage — a basic Forest card onto the battlefield tapped",
				ThenTail:      game.KeyedTail{Ref: nissasPilgrimageHand, Player: item.Controller, Source: ctx.Source(), N: hand},
			}.Apply(ctx)
		},
	})
}

// nissasPilgrimageHand is the hand half: up to N more basic Forest cards into
// the hand, then the shuffle. An on-disk key.
var nissasPilgrimageHand = game.RegisterTail("nissas-pilgrimage/forests-to-hand", func(g *game.Game, o game.TailOutcome) error {
	return g.SearchLibraryThenForEffect(game.SearchLibrarySpec{
		Player:  o.Tail.Player,
		Source:  o.Tail.Source,
		Pred:    effects.IsBasicLandOfAnySubtype("Forest"),
		Dest:    game.ZoneHand,
		Limit:   o.Tail.N,
		Reveal:  true,
		Shuffle: true,
		Reason:  "Nissa's Pilgrimage — basic Forest cards into your hand",
	})
})
