package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Through the Forest Gate — Sorcery {6}{G}{G}:
//
//	"Look at the top twenty cards of your library, put any number of
//	 land cards from among them onto the battlefield tapped, then
//	 shuffle. You gain 8 life."
//
// Genesis Wave's shape with a private look (LookAtTopOfLibraryForEffect:
// only the caster learns the cards, CR 701.20) and the lands entering
// tapped (PutFromLibraryOntoBattlefield.Tapped, Max 0 = "any number").
// Everything not put stays where it was and the whole library is
// shuffled afterwards, so the rider is a keyed tail rather than code
// after the prompt: it must run once the answer is in, and in every case
// (no lands to find, a forced answer, a declined one). The life gain
// follows the shuffle in the printed order.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "acfe54b3-10e3-4fdb-b874-d39fde96c40a",
		Name:         "Through the Forest Gate",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			looked := ctx.Game.LookAtTopOfLibraryForEffect(item.Controller, 20)
			return PutFromLibraryOntoBattlefield{
				Player:   item.Controller,
				Cards:    looked,
				Match:    Land(),
				Optional: true,
				Tapped:   true,
				Label:    "Through the Forest Gate — put any number of land cards onto the battlefield tapped",
				ThenTail: throughTheForestGateTail,
			}.Apply(ctx)
		},
	})
}

// throughTheForestGateTail is "then shuffle. You gain 8 life". On-disk
// key: never renamed, never reused.
var throughTheForestGateTail = PutFromLibraryTail("through-the-forest-gate/shuffle-and-gain-8", func(ctx *Context, res PutFromLibraryResult) error {
	if err := (ShuffleLibrary{Player: res.Player}).Apply(ctx); err != nil {
		return err
	}
	return GainLife{Player: res.Player, Amount: 8}.Apply(ctx)
})
