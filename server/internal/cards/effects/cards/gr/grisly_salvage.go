package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grisly Salvage — Instant {B}{G}:
//
//	"Reveal the top five cards of your library. You may put a creature or
//	 land card from among them into your hand. Put the rest into your
//	 graveyard."
//
// Satyr Wayfinder's shape with five cards and a creature-or-land pick: the
// five are revealed to every player, one optional pick, and everything not
// picked is put (not milled) into the graveyard.
//
// No simplification.
func init() {
	const label = "Grisly Salvage — reveal the top five cards, you may put a creature or land card into your hand and the rest into your graveyard"
	effects.Register(effects.Spec{
		OracleID:     "a89ebf63-63c9-42cd-98b4-ec852cd9b0a8",
		Name:         "Grisly Salvage",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			player := ctx.Controller()
			return effects.TakeFromLibraryToHand{
				Player:   player,
				Cards:    ctx.Game.RevealTopOfLibraryForEffect(player, ctx.Source(), 5, label),
				Match:    effects.Or(effects.Creature(), effects.Land()),
				Max:      1,
				Optional: true,
				Label:    label,
				Key:      effects.TakeRestGraveyard,
			}.Apply(ctx)
		},
	})
}
