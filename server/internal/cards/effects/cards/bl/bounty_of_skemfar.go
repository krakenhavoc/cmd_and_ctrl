package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bounty of Skemfar — Sorcery {2}{G} (Foundations Jumpstart):
//
//	"Reveal the top six cards of your library. You may put up to one land
//	 card from among them onto the battlefield tapped and up to one Elf
//	 card from among them into your hand. Put the rest on the bottom of
//	 your library in a random order."
//
// The six are revealed to everyone. The land is a PutFromLibraryOntoBattlefield
// (optional, one, tapped); its keyed tail is the second pick over the cards
// still in the library — an Elf card into the hand — whose own tail puts
// what is left on the bottom in a random order (takeRevealRestBottomRandom).
// A card that is both a land and an Elf can only be one of the two, since
// the land pick leaves the library first.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e4bfe005-7c4e-4178-ae41-a28eb2352ad1",
		Name:         "Bounty of Skemfar",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Lands: 1, Tutors: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			player := ctx.Controller()
			return effects.PutFromLibraryOntoBattlefield{
				Player:   player,
				Cards:    ctx.Game.RevealTopOfLibraryForEffect(player, ctx.Source(), 6, "Bounty of Skemfar — reveal the top six cards"),
				Match:    effects.Land(),
				Max:      1,
				Optional: true,
				Tapped:   true,
				Label:    "Bounty of Skemfar — you may put a land card onto the battlefield tapped",
				ThenTail: bountyOfSkemfarElf,
			}.Apply(ctx)
		},
	})
}

// bountyOfSkemfarElf is the second half: up to one Elf card of what is left
// into the hand, the rest to the bottom in a random order. An on-disk key.
var bountyOfSkemfarElf = effects.PutFromLibraryTail("bounty-of-skemfar/elf-then-rest",
	func(ctx *effects.Context, res effects.PutFromLibraryResult) error {
		return effects.TakeFromLibraryToHand{
			Player:   res.Player,
			Cards:    res.Rest,
			Match:    effects.HasSubtype("Elf"),
			Max:      1,
			Optional: true,
			Label:    "Bounty of Skemfar — you may put an Elf card into your hand",
			Key:      effects.TakeRevealRestBottomRandom,
		}.Apply(ctx)
	})
