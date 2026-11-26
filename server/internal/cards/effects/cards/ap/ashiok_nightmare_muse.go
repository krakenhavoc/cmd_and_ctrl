package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ashiok, Nightmare Muse — Legendary Planeswalker — Ashiok {3}{U}{B},
// loyalty 5 (Theros Beyond Death, #5514):
//
//	"+1: Create a 2/3 blue and black Nightmare creature token with "Whenever
//	 this token attacks or blocks, each opponent exiles the top two cards of
//	 their library."
//	 −3: Return target nonland permanent to its owner's hand, then that
//	 player exiles a card from their hand.
//	 −7: You may cast up to three spells from among face-up cards your
//	 opponents own from exile without paying their mana costs."
//
// The token is the printed Nightmare with its own attack-or-block trigger
// (once per block, Morkrut Necropod's pair). The −3 reads the owner before
// the return and asks them for the exile once the return has settled; an empty
// hand exiles nothing. The −7 grants one free, cast-only permission over every
// nonland, face-up card an opponent owns in exile with three casts to share
// (Doom Reigns Supreme's grant), offered during the resolution (CR 608.2g).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ebc9ff52-c914-4c48-9d35-0e4cd07741cb",
		Name:            "Ashiok, Nightmare Muse",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Create a 2/3 blue and black Nightmare creature token with \"Whenever this token attacks or blocks, each opponent exiles the top two cards of their library.\"",
				Cost:    effects.LoyaltyCost(1),
				Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1},
				Effect:  effects.Do(effects.CreateToken{Template: effects.OldB137AshiokNightmareToken(), N: 1}),
			},
			{
				Label:   "−3: Return target nonland permanent to its owner's hand, then that player exiles a card from their hand.",
				Cost:    effects.LoyaltyCost(-3),
				Targets: effects.TargetPermanent("target nonland permanent", effects.Nonland()),
				Effect:  effects.OldB137AshiokBounce,
			},
			{
				Label:  "−7: You may cast up to three spells from among face-up cards your opponents own from exile without paying their mana costs.",
				Cost:   effects.LoyaltyCost(-7),
				Effect: effects.OldB137AshiokCastFromExile,
			},
		},
	})
}
