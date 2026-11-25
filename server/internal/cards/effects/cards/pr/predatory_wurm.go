package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Predatory Wurm — Creature — Wurm {3}{G}, 4/4 (Core Set 2021, #5511):
//
//	"Vigilance
//	 This creature gets +2/+2 as long as you control a Garruk planeswalker."
//
// Vivien's Crocodile's conditional self pump over a planeswalker with the
// Garruk subtype.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "99ed6e18-7f14-4ea8-baee-ad6c081388ff",
		Name:            "Predatory Wurm",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Static: []game.StaticAbility{{
			Label:     "Gets +2/+2 as long as you control a Garruk planeswalker.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: effects.SelfWhileYouControlA(effects.And(effects.Planeswalker(), effects.Subtype("Garruk"))),
			Apply:     effects.PumpBy(2, 2),
		}},
	})
}
