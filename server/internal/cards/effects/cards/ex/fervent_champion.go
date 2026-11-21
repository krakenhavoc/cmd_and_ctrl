package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	attack := effects.WheneverThisAttacks(label, effects.FirstTargetGainsUntilEndOfTurn(1, label))
	attack.Targets = effects.Another(effects.TargetCreature("another target attacking Knight you control",
		effects.AttackingCreature(), effects.Subtype("Knight"), effects.YouControl()))
	effects.Register(effects.Spec{
		OracleID:        "8de50bda-1ef6-4ccf-9efc-fab10b7eb2e5",
		Name:            "Fervent Champion",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike", "haste"},
		Triggered:       []game.TriggeredAbility{attack},
		CostModifiers: []game.CostModifier{
			effects.EquipAbilitiesThatTargetThisCostLess(3,
				"Equip abilities you activate that target this creature cost {3} less to activate."),
		},
	})
}
