package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lorehold Pledgemage — Creature — Kor Shaman {1}{R/W}{R/W}, 2/2
// (Strixhaven: School of Mages, #5502):
//
//	"First strike
//	 Magecraft — Whenever you cast or copy an instant or sorcery spell,
//	 this creature gets +1/+0 until end of turn."
//
// First strike is the printed keyword; the magecraft is Eager First-Year's
// +1/+0 pump.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2c18d692-612c-4d43-b5dd-bd2bff37bbcb",
		Name:            "Lorehold Pledgemage",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			effects.Magecraft("Lorehold Pledgemage — gets +1/+0 until end of turn",
				effects.ThisGetsUntilEndOfTurn(1, 0, "Lorehold Pledgemage — magecraft pump")),
		},
	})
}
