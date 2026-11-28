package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zanam Djinn — Creature — Djinn {5}{U}, 5/6 (Invasion, S82 OLD #5616):
//
//	"Flying
//	 This creature gets -2/-2 as long as blue is the most common color among
//	 all permanents or is tied for most common."
//
// Flying is printed. The -2/-2 is the Mirage Djinns' layer 7c static
// (effects.SelfPumpWhileColorIsMostCommon): it counts the layered colours of
// every permanent on the battlefield, this creature's included (CR 105.2), and
// holds while blue is first or tied for first, so it comes and goes as the
// board changes.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "049fc248-4e46-43de-986d-fa023a3e8792",
		Name:            "Zanam Djinn",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{effects.SelfPumpWhileColorIsMostCommon("U", -2, -2)},
	})
}
