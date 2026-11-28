package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ruham Djinn — Creature — Djinn {5}{W}, 5/5 (Invasion, S82 OLD #5616):
//
//	"First strike
//	 This creature gets -2/-2 as long as white is the most common color among
//	 all permanents or is tied for most common."
//
// First strike is printed. The -2/-2 is the Mirage Djinns' layer 7c static
// (effects.SelfPumpWhileColorIsMostCommon): it counts the layered colours of
// every permanent on the battlefield, this creature's included (CR 105.2), and
// holds while white is first or tied for first, so it comes and goes as the
// board changes.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "16ad3b64-a87f-47e8-8e7b-b97bf084a3d2",
		Name:            "Ruham Djinn",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Static:          []game.StaticAbility{effects.SelfPumpWhileColorIsMostCommon("W", -2, -2)},
	})
}
