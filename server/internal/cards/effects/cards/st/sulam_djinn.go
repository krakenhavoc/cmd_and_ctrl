package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sulam Djinn — Creature — Djinn {5}{G}, 6/6 (Invasion, S82 OLD #5616):
//
//	"Trample
//	 This creature gets -2/-2 as long as green is the most common color among
//	 all permanents or is tied for most common."
//
// Trample is printed. The -2/-2 is the Mirage Djinns' layer 7c static
// (effects.SelfPumpWhileColorIsMostCommon): it counts the layered colours of
// every permanent on the battlefield, this creature's included (CR 105.2), and
// holds while green is first or tied for first, so it comes and goes as the
// board changes.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "01e311d2-8fe6-4e64-9575-3ef266985f0b",
		Name:            "Sulam Djinn",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static:          []game.StaticAbility{effects.SelfPumpWhileColorIsMostCommon("G", -2, -2)},
	})
}
