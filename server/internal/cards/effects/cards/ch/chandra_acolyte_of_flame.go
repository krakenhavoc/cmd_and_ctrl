package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chandra, Acolyte of Flame — Legendary Planeswalker — Chandra {1}{R}{R},
// loyalty 4 (Core Set 2020, #5518):
//
//	"0: Put a loyalty counter on each red planeswalker you control.
//	 0: Create two 1/1 red Elemental creature tokens. They gain haste.
//	 Sacrifice them at the beginning of the next end step.
//	 −2: You may cast target instant or sorcery card with mana value 3 or less
//	 from your graveyard. If that spell would be put into your graveyard, exile
//	 it instead."
//
// The first 0 adds one loyalty counter to every red planeswalker you control
// (Chandra included), each through the counter pipeline. The second 0 makes two
// Elementals through the keyed creation door and hands the ones that landed to a
// tail that gives them haste and schedules their sacrifice at the next end step
// (CR 603.7). The −2 targets a card in your graveyard, re-checked as it
// resolves, and offers the cast during the resolution (CR 608.2g) at the card's
// mana cost; the spell is exiled instead of going to a graveyard however it
// leaves the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "626e8acc-20da-496c-9a79-bfbf7529c01d",
		Name:            "Chandra, Acolyte of Flame",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 4,
		Activated: []effects.ActivatedAbility{
			{
				Label:  "0: Put a loyalty counter on each red planeswalker you control.",
				Cost:   effects.LoyaltyCost(0),
				Effect: effects.OldB155LoyaltyOnEachRedPlaneswalker,
			},
			{
				Label:   "0: Create two 1/1 red Elemental creature tokens. They gain haste. Sacrifice them at the beginning of the next end step.",
				Cost:    effects.LoyaltyCost(0),
				Purpose: game.Purpose{Tokens: 2, CreatureTokens: 2},
				Effect:  effects.OldB155ElementalsHasteSacrificed,
			},
			{
				Label: "−2: You may cast target instant or sorcery card with mana value 3 or less from your graveyard. If that spell would be put into your graveyard, exile it instead.",
				Cost:  effects.LoyaltyCost(-2),
				Targets: effects.TargetCardInGraveyard("target instant or sorcery card with mana value 3 or less from your graveyard",
					effects.Or(effects.Instant(), effects.Sorcery()), effects.ManaValueLE(3), effects.YouOwn()),
				Effect: effects.OldB155CastFromYourGraveyardExileIt("Chandra, Acolyte of Flame"),
			},
		},
	})
}
