package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Not Forgotten — Sorcery {1}{W} (Jumpstart 2022, #5948):
//
//	"Put target card from a graveyard on your choice of the top or bottom of
//	 its owner's library. Create a 1/1 white Spirit creature token with
//	 flying."
//
// Run Ashore's top-or-bottom tuck with the choice given to the spell's
// CONTROLLER rather than the card's owner ("your choice"), the card going
// to its owner's library. The Spirit is created whether or not the card is
// still there: a target that has become illegal fizzles the whole spell
// (CR 608.2b), but a legal one that cannot be moved (a commander offered the
// command zone, CR 903.9) still earns the token.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e0fa6d08-1808-4a29-b1e4-5cec9d31d798",
		Name:         "Not Forgotten",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCardInGraveyard("target card from a graveyard"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) > 0 && item.Targets[0].Kind == game.TargetCard {
				if err := (effects.PutIntoLibraryAtDepthOrBottom{
					Card: item.Targets[0].ID, Depth: 1, Chooser: item.Controller,
					Label: "Not Forgotten — put it on your choice of the top or bottom of its owner's library",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return effects.CreateToken{Template: effects.TokenCard("1/1 white Spirit with flying"), N: 1}.Apply(ctx)
		},
	})
}
