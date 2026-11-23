package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wrecking Ball Arm — Legendary Artifact — Equipment {2} (#3656):
//
//	"Equipped creature has base power and toughness 7/7 and can't be
//	 blocked by creatures with power 2 or less.
//	 Equip legendary creature {3}
//	 Equip {7}"
//
// The base-P/T set is a layer 7b static (SetAttachedBasePT), so counters
// and anthems still stack on the 7/7. The evasion is Kithkin Armor's
// block rule with the inverse comparison. Two equip abilities, the
// legendary-only one at {3} and the plain one at {7}, as Blackblade
// Reforged is written.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9f48cb32-46be-4c47-b558-2ea359efdd9b",
		Name:         "Wrecking Ball Arm",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.SetAttachedBasePT(7, 7),
		},
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedBy(effects.OnAttached(), effects.PowerLE(2), "creatures with power 2 or less"),
		},
		Activated: []effects.ActivatedAbility{
			effects.EquipOnlyAbility("Equip legendary creature {3}", "{3}",
				effects.TargetCreature("target legendary creature you control", effects.YouControl(), effects.Legendary())),
			effects.EquipAbility("{7}"),
		},
	})
}
