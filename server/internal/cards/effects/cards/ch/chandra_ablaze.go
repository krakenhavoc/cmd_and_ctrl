package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chandra Ablaze — Legendary Planeswalker — Chandra {4}{R}{R}, loyalty 5
// (Zendikar, #5583):
//
//	"+1: Discard a card. If a red card is discarded this way, Chandra deals 4
//	     damage to any target.
//	 −2: Each player discards their hand, then draws three cards.
//	 −7: Cast any number of red instant and/or sorcery cards from your
//	     graveyard without paying their mana costs."
//
// The +1's target is chosen as the ability goes on the stack; the discard is
// a choice from the hand, and only a red card (read off the card in the
// graveyard) pays the 4 damage. The −2's draws are a step after the discards
// (CR 608.2c). The −7 grants each red instant and sorcery card in the
// graveyard its own free cast, offered during the resolution (CR 608.2g).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3fae28e6-2ffc-460e-9dd3-e321e8a53ba9",
		Name:            "Chandra Ablaze",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Discard a card. If a red card is discarded this way, Chandra deals 4 damage to any target.",
				Cost:    effects.LoyaltyCost(1),
				Targets: effects.TargetAny(),
				Effect:  effects.OldB497ChandraDiscard,
			},
			{
				Label:  "−2: Each player discards their hand, then draws three cards.",
				Cost:   effects.LoyaltyCost(-2),
				Effect: effects.Do(effects.Instruction(effects.OldB497ChandraDiscardHands), effects.Instruction(effects.OldB497ChandraDrawThree)),
			},
			{
				Label:  "−7: Cast any number of red instant and/or sorcery cards from your graveyard without paying their mana costs.",
				Cost:   effects.LoyaltyCost(-7),
				Effect: effects.OldB497ChandraCastReds,
			},
		},
	})
}
