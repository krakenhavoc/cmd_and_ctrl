package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Halam Djinn — Creature — Djinn {5}{R}, 6/5 (Invasion, S82 OLD #5616):
//
//	"Haste
//	 This creature gets -2/-2 as long as red is the most common color among
//	 all permanents or is tied for most common."
//
// Haste is printed. The -2/-2 is the Mirage Djinns' layer 7c static
// (effects.SelfPumpWhileColorIsMostCommon): it counts the layered colours of
// every permanent on the battlefield, this creature's included (CR 105.2), and
// holds while red is first or tied for first, so it comes and goes as the
// board changes.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fb85d8d1-6938-48bf-9a6d-d4616b05b4f6",
		Name:            "Halam Djinn",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Static:          []game.StaticAbility{effects.SelfPumpWhileColorIsMostCommon("R", -2, -2)},
	})
}
