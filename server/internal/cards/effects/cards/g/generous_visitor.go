package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Generous Visitor — Creature — Spirit {G}, 1/1 (Kamigawa: Neon Dynasty,
// S82):
//
//	"Whenever you cast an enchantment spell, put a +1/+1 counter on target
//	 creature."
//
// An enchantment-spell cast watcher with a creature target, any creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0629747e-8dd8-4150-95ee-9d860264ccb0",
		Name:         "Generous Visitor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverYouCast(effects.Enchantment(), "Generous Visitor — put a +1/+1 counter on target creature", effects.PlusOneCounterOnChosenTarget),
				effects.TargetCreature("target creature")),
		},
	})
}
