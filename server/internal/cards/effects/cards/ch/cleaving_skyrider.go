package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cleaving Skyrider — Creature — Human Warrior {2}{W}, 2/2 (Dominaria
// United):
//
//	"Flash
//	 Kicker {2}{R}
//	 Flying
//	 When this creature enters, if it was kicked, it deals X damage to any
//	 target, where X is the number of attacking creatures."
//
// Archangel of Wrath's kicked enters trigger (an intervening "if", CR 603.4)
// with Galadhrim Ambush's count of every attacking creature, whoever controls
// it. X is read once as the ability resolves (CR 608.2h) and the Skyrider is
// the damage source; a target gone by resolution makes it do nothing
// (CR 608.2b), and X = 0 deals no damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "eb1ab332-5549-4ff4-8afa-b385a236fb9b",
		Name:            "Cleaving Skyrider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		OptionalCosts:   []game.AdditionalCost{effects.Kicker("{2}{R}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventETB, effects.AllOf(effects.Self, effects.ThisKickedAtLeast(1)),
					"Cleaving Skyrider — kicked, it deals X damage to any target, where X is the number of attacking creatures",
					effects.SourceDealsComputedDamageToEachLegalTarget(func(ctx *effects.Context) int {
						return len(effects.AttackingCreatures(ctx.Game))
					})),
				effects.TargetAny()),
		},
	})
}
