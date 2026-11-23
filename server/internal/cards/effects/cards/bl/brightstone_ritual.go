package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Brightstone Ritual — Instant {R} (Onslaught):
//
//	"Add {R} for each Goblin on the battlefield."
//
// Mana Geyser's counted AddMana over every player's Goblins (Airborne
// Aid's table-wide count), read as the spell resolves. "Goblin" is a word
// of the card's text (CR 612.2), so a changeling counts. With no Goblin
// it adds nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "08e90e85-4103-4acb-a8a7-e1329b460aa7",
		Name:         "Brightstone Ritual",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			n := effects.B42CountCreaturesWithSubtype(ctx.Game, ctx.Game.TextReaderForEffect(item).TextSubtype("Goblin"))
			if n == 0 {
				return nil
			}
			return effects.AddMana{Player: item.Controller, Amount: effects.ManaN(n, "R")}.Apply(ctx)
		},
	})
}
