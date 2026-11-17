package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "eb1ab332-5549-4ff4-8afa-b385a236fb9b",
		Name:            "Cleaving Skyrider",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		OptionalCosts:   []game.AdditionalCost{Kicker("{2}{R}")},
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventETB, AllOf(Self, ThisKickedAtLeast(1)),
					"Cleaving Skyrider — kicked, it deals X damage to any target, where X is the number of attacking creatures",
					sourceDealsComputedDamageToEachLegalTarget(func(ctx *Context) int {
						return len(AttackingCreatures(ctx.Game))
					})),
				TargetAny()),
		},
	})
}
