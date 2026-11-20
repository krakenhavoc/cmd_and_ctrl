package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gold-Forged Thopteryx — Artifact Creature — Dinosaur Thopter {W}{U},
// 1/3 (March of the Machine: The Aftermath):
//
//	"Flying, lifelink
//	 Each legendary permanent you control has ward {2}."
//
// Flying and lifelink are printed keywords. The ward is WardGranted
// (Bronze Guardian's shape) over every legendary permanent its controller
// controls, read through the layers; the Thopteryx is not legendary, so
// it does not ward itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c27e2d6f-498c-42d8-aa1a-ac6402bf3f9a",
		Name:            "Gold-Forged Thopteryx",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.WardGranted(effects.WardMana("{2}"), "Gold-Forged Thopteryx — each legendary permanent you control has ward {2}",
				func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.Controller == source.Controller && target.IsLegendary()
				}),
		},
	})
}
