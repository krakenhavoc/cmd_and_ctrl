package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ultron, Machine Overlord — Legendary Artifact Creature — Robot Villain
// {5}, 4/4 (Marvel Super Heroes Commander, #4455):
//
//	"Flying
//	 Other Robots and Constructs you control get +2/+2."
//
// A two-tribe lord (TribalAnthem over a TribeFilter naming Robot and
// Construct, others, yours): a creature that is both gets +2/+2 once
// (the filter matches the creature, not each type). Layer 7c.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "19fd953c-b6d4-457f-b85f-b681c5b0e341",
		Name:            "Ultron, Machine Overlord",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Robot", "Construct"}, Others: true, YoursOnly: true}, 2, 2),
		},
	})
}
