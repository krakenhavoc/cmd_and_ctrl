package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Picklock Prankster // Free the Fae — Creature — Faerie Rogue {1}{U}, 1/3
// // Instant — Adventure {1}{U} (Wilds of Eldraine, #4471):
//
//	Picklock Prankster — "Flying, vigilance"
//	Free the Fae — "Mill four cards. Then put an instant, sorcery, or
//	 Faerie card from among the milled cards into your hand. (Then exile
//	 this card. You may cast the creature later from exile.)"
//
// Both faces register under the Adventure keyspace (the bare oracle ID and
// "#1"). Free the Fae is Eerie Gravestone's mill-then-take with the pick
// mandatory ("put", not "you may put"): only the cards this mill put into
// the graveyard are offered, and a milled card that matches is always
// taken (a floor of one). Nothing matching milled is simply no pick. A
// card's Faerie type is read as the card's text gives it (CR 612.2).
//
// No simplification.
const picklockPranksterOracleID = "22d82f4c-2896-477e-8d20-8fe43acd4ba2"

func init() {
	Register(Spec{
		OracleID:        picklockPranksterOracleID,
		Name:            "Picklock Prankster",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
	})
	Register(Spec{
		OracleID:     picklockPranksterOracleID + "#1",
		Name:         "Free the Fae",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return MillToZone{N: 4, ThenTail: ctx.ItemTail(freeTheFaeMilled)}.Apply(ctx)
		},
	})
}

// freeTheFaeMilled is the pick after the mill. An on-disk key: never
// renamed, never reused.
var freeTheFaeMilled = CardTail("picklock-prankster/after-mill", func(ctx *Context, o game.TailOutcome) error {
	return takeOneFromAmongThem(ctx, o.IDs, Or(Instant(), Sorcery(), OfSubtype("Faerie")),
		"Free the Fae — put an instant, sorcery, or Faerie card from among the milled cards into your hand")
})
