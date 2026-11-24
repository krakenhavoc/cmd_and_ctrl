package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// General Ferrous Rokiric — Legendary Creature — Human Soldier {1}{R}{W}, 3/1
// (Modern Horizons 2):
//
//	"Hexproof from monocolored
//	 Whenever you cast a multicolored spell, create a 4/4 red and white Golem
//	 artifact creature token."
//
// Hexproof from monocolored is the engine's quality (CR 702.11d). The trigger
// reads the spell's colours as it was cast, so a hybrid or two-colour spell
// counts and a colorless or one-colour one does not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "64c15098-f893-44f2-bf7f-162db8ab80e9",
		Name:            "General Ferrous Rokiric",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.HexproofFromMonocolored},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Multicolored(), "General Ferrous Rokiric — create a 4/4 red and white Golem artifact creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("4/4 red and white Golem artifact"), N: 1})),
		},
	})
}
