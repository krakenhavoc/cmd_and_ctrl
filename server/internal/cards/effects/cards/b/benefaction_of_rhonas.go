package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Benefaction of Rhonas — {2}{G} Sorcery (Amonkhet, S82 old batch 247):
//
//	"Reveal the top five cards of your library. You may put a creature card
//	 and/or an enchantment card from among them into your hand. Put the rest
//	 into your graveyard."
//
// Two slots of one card each (Gift of the Gargantuan's shape): a card that is
// both a creature and an enchantment fills only one of them. The five are
// revealed, and everything not taken goes to the graveyard.
//
// No simplification.
func init() {
	const label = "Benefaction of Rhonas — reveal the top five cards, you may put a creature card and/or an enchantment card into your hand and the rest into your graveyard"
	effects.Register(effects.Spec{
		OracleID:     "02da2cbf-50d6-4c37-8a9f-40db1f7954be",
		Name:         "Benefaction of Rhonas",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			player := ctx.Controller()
			return effects.TakeFromLibraryToHand{
				Player:   player,
				Cards:    ctx.Game.RevealTopOfLibraryForEffect(player, ctx.Source(), 5, label),
				Optional: true,
				Label:    label,
				Key:      benefactionOfRhonasTake,
			}.Apply(ctx)
		},
	})
}

var benefactionOfRhonasTake = effects.RegisterTake("choose-cards/s82-old-b247-benefaction-of-rhonas-take",
	effects.TakeDef{
		Slots: []effects.TakeSlot{
			{Label: "a creature card", Match: effects.Creature(), Max: 1},
			{Label: "an enchantment card", Match: effects.Enchantment(), Max: 1},
		},
		Then: effects.TakeRestIntoGraveyard,
	})
