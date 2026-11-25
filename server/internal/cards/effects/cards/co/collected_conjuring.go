package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Collected Conjuring — Sorcery {2}{U}{R} (Modern Horizons):
//
//	"Exile the top six cards of your library. You may cast up to two sorcery
//	 spells with mana value 3 or less from among them without paying their
//	 mana costs. Put the exiled cards not cast this way on the bottom of your
//	 library in a random order."
//
// The six cards are exiled (not milled), then one permission over the
// sorceries of mana value 3 or less among them, with two casts to spend
// (Plargg and Nassari's), lets the controller cast up to two of them for {0};
// the casts are made during the resolution (CR 608.2g), each offered in turn
// through the resolution_cast prompt until two are made or the player
// declines. An {X} in a cost is 0 (CR 107.3b) and a mandatory additional cost
// is still paid. Whatever is answered, the cards still in exile then go to the
// bottom in a random order. A commander the exile would have moved to the
// command zone is not among them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a0a79e9e-38cb-45d8-a7ec-7cf10ef7e5fc",
		Name:         "Collected Conjuring",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.MillToZone{
				Player:   ctx.Controller(),
				N:        6,
				To:       game.ZoneExile,
				ThenTail: ctx.ItemTail(collectedConjuringExiled),
			}.Apply(ctx)
		},
	})
}

// collectedConjuringExiled is the clause after the exile, told the cards
// exiled. An on-disk key.
var collectedConjuringExiled = effects.CardTail("s82-old-b162/collected-conjuring-exiled", func(ctx *effects.Context, o game.TailOutcome) error {
	return effects.S82OldB162CastUpToTwoSorceriesThenBottom(ctx.Game, ctx.Controller(), ctx.Source(), "Collected Conjuring", o.IDs)
})
