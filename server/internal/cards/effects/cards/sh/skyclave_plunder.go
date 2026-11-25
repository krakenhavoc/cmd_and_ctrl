package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Skyclave Plunder — Sorcery {4}{U} (Zendikar Rising, #5508):
//
//	"Look at the top X cards of your library, where X is three plus the
//	 number of creatures in your party. Put three of those cards into your
//	 hand and the rest on the bottom of your library in a random order. (Your
//	 party consists of up to one each of Cleric, Rogue, Warrior, and Wizard.)"
//
// X is the party (CR 700.8) as the spell resolves. A look, so only the caster
// sees the cards, and mandatory: exactly three go to the hand (all of them if
// the library holds fewer, CR 401.5), nothing is revealed, and the rest go to
// the bottom in a random order.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d2d5f9b8-7caf-42c8-934b-3f4b81468914",
		Name:         "Skyclave Plunder",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 3},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			g, you := ctx.Game, ctx.Controller()
			return effects.TakeFromLibraryToHand{
				Player: you,
				Cards:  g.LookAtTopOfLibraryForEffect(you, 3+effects.OldB096PartyOf(g, you)),
				Max:    3,
				Label:  "Skyclave Plunder — put three of those cards into your hand",
				Key:    effects.TakeRestBottomRandom,
			}.Apply(ctx)
		},
	})
}
