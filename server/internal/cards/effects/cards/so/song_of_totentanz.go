package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Song of Totentanz — Sorcery {X}{R} (Wilds of Eldraine, #4471):
//
//	"Create X 1/1 black Rat creature tokens with "This token can't
//	 block." Creatures you control gain haste until end of turn."
//
// The Rats are Twisted Sewer-Witch's catalog token, whose own text is
// "This token can't block", X of them under the caster. The haste is
// Crashing Drawbridge's GrantKeywordUntilEOT over the creatures you
// control as it resolves (CR 611.2c), after the Rats have entered, so the
// new Rats have haste too. The haste never reads X, so X=0 is a real
// play and XMatters stays unset (the allowlist says so).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5f4d0c87-7d59-41d5-9910-86765a0c144b",
		Name:         "Song of Totentanz",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := (effects.CreateToken{Controller: ctx.Controller(), Template: effects.CantBlockRatToken(), N: ctx.X()}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantKeywordUntilEOT{
				Match:    effects.And(effects.Creature(), effects.YouControl()),
				Keywords: []string{"haste"},
				Label:    "Song of Totentanz — haste",
			}.Apply(ctx)
		},
	})
}
