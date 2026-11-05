package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Branching Timeline — Sorcery {6}{G}{G} (Star Trek, #4973):
//
//	"Strange new worlds — Reveal the top X cards of your library, where X
//	 is the number of differently named lands you control. Put all
//	 permanent cards revealed this way onto the battlefield tapped and the
//	 rest on the bottom of your library in a random order."
//
// "Strange new worlds" is an ability word (CR 207.2c). X is counted as
// the spell resolves (CR 608.2h), each land name once. "All permanent
// cards" is not a choice, so there is no prompt: Animist's Awakening's
// reveal and put, with no filter beyond the permanent-card rule, every
// card entering tapped in one simultaneous batch, and the rest go to the
// bottom in a random order. An X of 0 reveals nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6e13c619-706d-4c77-bc7d-187df0fd054a",
		Name:         "Branching Timeline",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			x := socB23DifferentlyNamedLands(ctx)
			if x <= 0 {
				return nil
			}
			revealed := ctx.Game.RevealTopOfLibraryForEffect(item.Controller, ctx.Source(), x,
				"Branching Timeline — revealed from the top of the library")
			return PutFromLibraryOntoBattlefield{
				Player:   item.Controller,
				Cards:    revealed,
				All:      true,
				Tapped:   true,
				Label:    "Branching Timeline — put all permanent cards revealed this way onto the battlefield tapped",
				ThenTail: PutRestOnBottomInRandomOrderTail,
			}.Apply(ctx)
		},
	})
}
