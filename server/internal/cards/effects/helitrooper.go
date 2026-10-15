package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Helitrooper — Creature — Human Soldier {1}{W}, 1/2 (#3656, #3778):
//
//	"Flying
//	 Whenever this creature attacks, another target attacking creature
//	 gains flying until end of turn.
//	 Equip abilities you activate that target this creature cost {2}
//	 less to activate."
//
// The discount reads the equip bit and the announced target off the
// activation's cost query (EquipAbilitiesThatTargetThisCostLess), so
// it applies to any Equipment's equip ability aimed at Helitrooper and
// to no other ability.
//
// No simplification.
func init() {
	const label = "Helitrooper — another attacking creature gains flying"
	attack := WheneverThisAttacks(label, firstTargetGainsUntilEndOfTurn(0, label, "flying"))
	attack.Targets = Another(TargetCreature("another target attacking creature", AttackingCreature()))
	Register(Spec{
		OracleID:        "a2900c79-8e55-49f9-b548-5890a753932e",
		Name:            "Helitrooper",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{attack},
		CostModifiers: []game.CostModifier{
			EquipAbilitiesThatTargetThisCostLess(2,
				"Equip abilities you activate that target this creature cost {2} less to activate."),
		},
	})
}
