package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Commune with Dinosaurs — Sorcery {G} (Marvel Super Heroes Commander):
//
//	"Look at the top five cards of your library. You may reveal a Dinosaur
//	 or land card from among them and put it into your hand. Put the rest
//	 on the bottom of your library in any order."
//
// Oath of Nissa's dig (TakeFromLibraryToHand with takeRevealRestBottomAny):
// the pick is optional, at most one card, revealed, and the rest go to the
// bottom in an order the player chooses.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "27ca27f0-95a5-4292-b30b-1db3e9d639f8",
		Name:         "Commune with Dinosaurs",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			player := ctx.Controller()
			return effects.TakeFromLibraryToHand{
				Player:   player,
				Cards:    ctx.Game.LookAtTopOfLibraryForEffect(player, 5),
				Match:    effects.Or(effects.HasSubtype(ctx.TextSubtype("Dinosaur")), effects.Land()),
				Max:      1,
				Optional: true,
				Label:    "Commune with Dinosaurs — reveal a Dinosaur or land card and put it into your hand",
				Key:      effects.TakeRevealRestBottomAny,
			}.Apply(ctx)
		},
	})
}
