package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "0629747e-8dd8-4150-95ee-9d860264ccb0",
		Name:         "Generous Visitor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverYouCast(Enchantment(), "Generous Visitor — put a +1/+1 counter on target creature", PlusOneCounterOnChosenTarget),
				TargetCreature("target creature")),
		},
	})
}
