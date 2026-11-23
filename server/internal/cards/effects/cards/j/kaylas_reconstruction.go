package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kayla's Reconstruction — Sorcery {X}{W}{W}{W} (The Brothers' War, S82):
//
//	"Look at the top seven cards of your library. Put up to X artifact and/or
//	 creature cards with mana value 3 or less from among them onto the
//	 battlefield. Put the rest on the bottom of your library in a random
//	 order."
//
// The announced X caps the picks; with X of 0 nothing may be picked and the
// whole look goes to the bottom. The cards enter under your control and the
// rest is bottomed in the game's seeded random order.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "43bae4b3-57e8-4d9a-9d1b-9805b0e0fff5",
		Name:         "Kayla's Reconstruction",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			player := ctx.Controller()
			looked := ctx.Game.LookAtTopOfLibraryForEffect(player, 7)
			x := ctx.X()
			if x <= 0 {
				// Nothing may be put onto the battlefield: the look just goes
				// to the bottom.
				return effects.PutFromLibraryOntoBattlefield{
					Player:   player,
					Cards:    looked,
					Match:    func(*game.Game, uuid.UUID, game.Card) bool { return false },
					All:      true,
					ThenTail: effects.PutRestOnBottomInRandomOrderTail,
				}.Apply(ctx)
			}
			return effects.PutFromLibraryOntoBattlefield{
				Player:   player,
				Cards:    looked,
				Match:    effects.And(effects.Or(effects.Artifact(), effects.Creature()), effects.ManaValueLE(3)),
				Max:      x,
				Optional: true,
				Label:    "Kayla's Reconstruction — put up to X artifact and/or creature cards with mana value 3 or less onto the battlefield",
				ThenTail: effects.PutRestOnBottomInRandomOrderTail,
			}.Apply(ctx)
		},
	})
}
