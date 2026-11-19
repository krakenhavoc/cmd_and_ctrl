package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Heartstabber Mosquito — Creature — Insect {3}{B}, 2/2 (Foundations
// Jumpstart, BNS batch 32):
//
//	"Kicker {2}{B} (You may pay an additional {2}{B} as you cast this spell.)
//	 Flying
//	 When this creature enters, if it was kicked, destroy target creature."
//
// Kor Sanctifiers' kicked enters trigger (ThisKickedAtLeast, an intervening
// "if" read off the permanent's carried kicker record, so an unkicked
// Mosquito asks for no target) over any creature, your own included. A
// target that left in response is skipped (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f8fca88f-5eab-4854-913c-44261720a4d2",
		Name:            "Heartstabber Mosquito",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		OptionalCosts:   []game.AdditionalCost{Kicker("{2}{B}")},
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventETB, AllOf(Self, ThisKickedAtLeast(1)),
					"Heartstabber Mosquito — kicked, destroy target creature",
					destroyFirstLegalCardTarget),
				TargetCreature("target creature")),
		},
	})
}
