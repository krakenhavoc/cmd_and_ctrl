package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Moogles' Valor — Instant {3}{W}{W} (EDHREC rank 4462):
//
//	"For each creature you control, create a 1/2 white Moogle creature
//	 token with lifelink. Then creatures you control gain
//	 indestructible until end of turn."
//
// A five-mana instant that doubles your board and then makes the
// whole thing survive a Wrath. The order printed on the card is the
// order it matters in: the tokens are created FIRST, so they are on
// the battlefield when the indestructible is handed out and they get
// it too — which is what turns this from a combat trick into a
// counterspell for board wipes.
//
// Both halves are snapshots taken as the spell resolves:
//
//   - The token COUNT is the creatures you control at that moment.
//     A creature killed in response really does cost you a Moogle
//     (CR 608.2).
//   - The indestructible set is CR 611.2c: every creature you
//     control as the spell finishes, the fresh Moogles included, and
//     nothing that enters afterwards.
//
// The counting and the granting are therefore two separate walks of
// the battlefield, in that order, and the second one is
// GrantKeywordUntilEOT's own Match snapshot rather than a reuse of
// the first list.
//
// Indestructible has been live since S25 (#77) and covers the mass
// destroy path since S30 (#470), so "survives a Wrath" is true of
// the implementation. Lifelink on the tokens is a canonical keyword
// and is real.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9627ed9a-5ade-412d-bf8d-f0705b1c9405",
		Name:         "Moogles' Valor",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			n := b04CreaturesControlled(ctx.Game, ctx.Controller())
			if n > 0 {
				if err := (CreateToken{
					Controller: ctx.Controller(),
					Template:   TokenCard("1/2 white Moogle with lifelink"),
					N:          n,
				}).Apply(ctx); err != nil {
					return err
				}
			}
			controller := ctx.Controller()
			return GrantKeywordUntilEOT{
				Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
					return c.IsCreature() && c.Controller == controller
				},
				Keywords: []string{"indestructible"},
				Label:    "Moogles' Valor — creatures you control gain indestructible",
			}.Apply(ctx)
		},
	})
}
