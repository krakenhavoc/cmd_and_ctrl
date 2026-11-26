package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gather the Pack — Sorcery {1}{G} (Magic Origins, #5549):
//
//	"Reveal the top five cards of your library. You may put a creature card
//	 from among them into your hand. Put the rest into your graveyard.
//	 Spell mastery — If there are two or more instant and/or sorcery cards in
//	 your graveyard, put up to two creature cards from among the revealed
//	 cards into your hand instead of one."
//
// The five cards are revealed to the table, then the caster takes a creature
// card (up to two when spell mastery holds as the spell resolves, so it does
// not count itself) and the rest, whatever they are, go to the graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8f31d870-872d-4671-80d1-8d2ecd2c38d2",
		Name:         "Gather the Pack",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			player := ctx.Controller()
			max := 1
			if effects.SpellMastery(ctx.Game, player) {
				max = 2
			}
			revealed := ctx.Game.RevealTopOfLibraryForEffect(player, ctx.Source(), 5,
				"Gather the Pack — revealed from the top of the library")
			return effects.TakeFromLibraryToHand{
				Player:   player,
				Cards:    revealed,
				Match:    effects.Creature(),
				Max:      max,
				Optional: true,
				Label:    "Gather the Pack — put a creature card (two with spell mastery) into your hand and the rest into your graveyard",
				Key:      gatherThePackTake,
			}.Apply(ctx)
		},
	})
}

var gatherThePackTake = effects.RegisterTake("choose-cards/gather-the-pack-take",
	effects.TakeDef{Then: effects.TakeRestIntoGraveyard})
