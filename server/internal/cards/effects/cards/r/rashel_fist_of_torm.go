package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rashel, Fist of Torm — Legendary Creature — Human Knight {2}{W}{W}, 2/4
// (Secret Lair Drop):
//
//	"Double strike
//	 Auras you control have exalted. (Whenever a creature you control attacks
//	 alone, it gets +1/+1 until end of turn for each instance of exalted among
//	 permanents you control.)"
//
// Double strike is printed data. The grant is a layer-6 exalted to every Aura
// its controller has, each its own instance (CR 113.2c, 702.83): the engine's
// exalted trigger is derived from a permanent's ability list, so an Aura with it
// triggers for a lone attacker of its controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c0c39f80-cf12-482e-a7c4-757db85125b0",
		Name:            "Rashel, Fist of Torm",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		Static:          []game.StaticAbility{effects.KeywordGrant(effects.S82OldB024AurasYouControl, game.KeywordExalted)},
	})
}
