package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gatekeeper Gargoyle — Artifact Creature — Gargoyle {6}, 3/3 (Guilds of
// Ravnica):
//
//	"Flying
//	 This creature enters with a +1/+1 counter on it for each Gate you
//	 control."
//
// Dakkon's counted entry replacement over the Gates its controller controls
// as it enters (CR 614.1c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a175de44-3a7e-4165-8f9a-8173c0df13d6",
		Name:            "Gatekeeper Gargoyle",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements: []game.ReplacementEffect{
			effects.B19EntersWithCountersCounted(game.CounterPlusOne, func(g *game.Game, src *game.Card) int {
				return effects.CountControlled(g, src.Controller, func(c game.Card) bool { return c.HasSubtype("Gate") })
			}, "Gatekeeper Gargoyle: enters with a +1/+1 counter on it for each Gate you control"),
		},
	})
}
