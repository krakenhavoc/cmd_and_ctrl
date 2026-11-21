package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// God-Pharaoh's Faithful — Creature — Human Wizard {W}, 0/4 (Hour of
// Devastation, #5534):
//
//	"Whenever you cast a blue, black, or red spell, you gain 1 life."
//
// A cast trigger over the spell's colours (Balefire Liege's shape): one
// spell of two of those colours is one trigger, not two.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3f86e6c6-fca5-43e9-88cc-4546cdd5f0ec",
		Name:         "God-Pharaoh's Faithful",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Or(effects.OfColor("U"), effects.OfColor("B"), effects.OfColor("R")), "God-Pharaoh's Faithful — you gain 1 life", effects.Do(effects.GainLife{Amount: 1})),
		},
	})
}
