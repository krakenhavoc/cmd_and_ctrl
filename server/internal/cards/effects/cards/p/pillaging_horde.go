package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pillaging Horde — Creature — Human Barbarian {2}{R}{R}, 5/5 (Portal, S82 OLD
// batch 783, #5630):
//
//	"When this creature enters, sacrifice it unless you discard a card at
//	 random."
//
// With an empty hand it is sacrificed at once; otherwise its controller chooses
// between discarding a card at random (the engine's keyed random pick) and
// sacrificing it, and only the same permanent that entered is sacrificed (CR
// 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ad07cb55-cc4f-40be-b16a-bd3d3ca94249",
		Name:         "Pillaging Horde",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Pillaging Horde — sacrifice it unless you discard a card at random", effects.S82OldB783SacrificeUnlessDiscardAtRandom("Pillaging Horde")),
		},
	})
}
