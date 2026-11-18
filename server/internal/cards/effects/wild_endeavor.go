package effects

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wild Endeavor — Sorcery {4}{G}{G} (Adventures in the Forgotten Realms):
//
//	"Roll two d4 and choose one result. Create a number of 3/3 green Beast
//	 creature tokens equal to that result. Then search your library for a
//	 number of basic land cards equal to the other result, put them onto
//	 the battlefield tapped, then shuffle."
//
// The shared two-dice shape is afrEndeavorTail. The search has a stated
// quality, so the searcher may find fewer (CR 701.23b).
func init() {
	Register(Spec{
		OracleID: "2c4ee0df-e720-475b-aa04-e6394f3d708f", Name: "Wild Endeavor", Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return RollDice{Sides: 4, N: 2, Then: wildEndeavorRolled}.Apply(ctx)
		},
	})
}

var wildEndeavorRolled = afrEndeavorTail("dice/wild-endeavor", "Wild Endeavor",
	func(chosen, other int) string {
		return strconv.Itoa(chosen) + " Beasts, search for " + strconv.Itoa(other) + " basic lands"
	},
	func(ctx *Context, chosen, other int) error {
		if err := (CreateToken{Template: TokenCard("3/3 green Beast"), N: chosen}).Apply(ctx); err != nil {
			return err
		}
		return b26SearchBasicsOntoBattlefieldTapped(ctx.Game, ctx.Item, other, "Wild Endeavor — search for basic land cards")
	})
