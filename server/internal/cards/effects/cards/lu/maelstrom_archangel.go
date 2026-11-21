package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Maelstrom Archangel — Creature — Angel {W}{U}{B}{R}{G}, 5/5:
//
//	"Flying
//	 Whenever this creature deals combat damage to a player, you may cast a
//	 spell from your hand without paying its mana cost."
//
// The free cast is the shared hand-cast offer (free_hand_cast.go): a
// pass-closed {0} permission over the non-land cards in the hand, made during
// the trigger's resolution (CR 608.2g, CR 601.2). Declining leaves the cards
// in hand; a land is not a spell and is never offered.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ac50ef98-1791-4dd4-9c1f-8cea7db7ade5",
		Name:            "Maelstrom Archangel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Maelstrom Archangel — you may cast a spell from your hand without paying its mana cost",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					effects.OfferFreeHandCast(ctx, ctx.Controller(), maelstromArchangelFreeCast, "a spell", effects.AnySpell)
					return nil
				}))),
		},
	})
}

var maelstromArchangelFreeCast = effects.RegisterFreeHandCastOffer("choose-cards/maelstrom-archangel-cast", "Maelstrom Archangel")
