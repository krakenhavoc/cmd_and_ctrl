package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wurmquake — Sorcery {4}{G}{G}:
//
//	"Corrupted — Create an X/X green Phyrexian Wurm creature token with
//	 trample and toxic 1, where X is the amount of mana spent to cast this
//	 spell. Then for each opponent with three or more poison counters, you
//	 create another one of those tokens.
//	 Flashback {8}{G}{G}"
//
// "Corrupted" is an ability word with no rules meaning (CR 207.2c): the
// first token is made whatever the poison counts, and each opponent with
// three or more poison counters, counted as the spell resolves, adds one
// more. X is the payment record's total (Molten Note): six from hand, ten
// from the graveyard with its flashback cost. A spell cast for no mana
// makes 0/0 tokens, which die as a state-based action (CR 704.5f). Every
// token has the same X.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "0eeb9a62-1b36-42f5-8820-3331fa5ca3a6",
		Name:             "Wurmquake",
		Completeness:     CompletenessFull,
		Purpose:          game.Purpose{Tokens: 1},
		ReadsManaSpent:   game.ReadsOwnPayment,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{8}{G}{G}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			x := ctx.ManaSpent().Total()
			n := 1 + s82OneB16CorruptedOpponents(ctx.Game, ctx.Controller())
			tok := ctx.TextToken(TokenCard("0/0 green Phyrexian Wurm with trample and toxic 1"))
			tok.Power, tok.Toughness = x, x
			return ctx.Game.CreateTokenForEffect(ctx.Controller(), tok, n)
		},
	})
}
