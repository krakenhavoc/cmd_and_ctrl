package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Commune with the Gods — Sorcery {1}{G} (Eternal Masters, S82 OLD, #5542):
//
//	"Reveal the top five cards of your library. You may put a creature or
//	 enchantment card from among them into your hand. Put the rest into your
//	 graveyard."
//
// Tapping at the Window's take with the reveal and the graveyard rest, over
// five revealed cards, at most one creature-or-enchantment card, optional.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6042a806-6e6a-4f1d-9563-5932a17f32f9",
		Name:         "Commune with the Gods",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			player := ctx.Controller()
			return effects.TakeFromLibraryToHand{
				Player:   player,
				Cards:    ctx.Game.RevealTopOfLibraryForEffect(player, ctx.Source(), 5, "Commune with the Gods — reveal the top five cards"),
				Match:    effects.Or(effects.Creature(), effects.Enchantment()),
				Max:      1,
				Optional: true,
				Label:    "Commune with the Gods — you may put a creature or enchantment card into your hand",
				Key:      effects.MidB03TakeRevealRestGraveyard,
			}.Apply(ctx)
		},
	})
}
