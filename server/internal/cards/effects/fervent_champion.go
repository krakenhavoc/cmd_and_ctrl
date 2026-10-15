package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fervent Champion — Creature — Human Knight {R}, 1/1 (#3778):
//
//	"First strike, haste
//	 Whenever this creature attacks, another target attacking Knight you
//	 control gets +1/+0 until end of turn.
//	 Equip abilities you activate that target this creature cost {3}
//	 less to activate."
//
// The discount is the equip-bit cost modifier; a reduction only takes
// generic mana off (CR 118.7a), so an equip cost with coloured symbols
// keeps them.
//
// No simplification.
func init() {
	const label = "Fervent Champion — another attacking Knight gets +1/+0"
	attack := WheneverThisAttacks(label, firstTargetGainsUntilEndOfTurn(1, label))
	attack.Targets = Another(TargetCreature("another target attacking Knight you control",
		AttackingCreature(), Subtype("Knight"), YouControl()))
	Register(Spec{
		OracleID:        "8de50bda-1ef6-4ccf-9efc-fab10b7eb2e5",
		Name:            "Fervent Champion",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike", "haste"},
		Triggered:       []game.TriggeredAbility{attack},
		CostModifiers: []game.CostModifier{
			EquipAbilitiesThatTargetThisCostLess(3,
				"Equip abilities you activate that target this creature cost {3} less to activate."),
		},
	})
}
