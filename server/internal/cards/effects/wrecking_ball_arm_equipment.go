package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "9f48cb32-46be-4c47-b558-2ea359efdd9b",
		Name:         "Wrecking Ball Arm",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			SetAttachedBasePT(7, 7),
		},
		BlockRules: []game.BlockRule{
			CantBeBlockedBy(OnAttached(), PowerLE(2), "creatures with power 2 or less"),
		},
		Activated: []ActivatedAbility{
			EquipOnlyAbility("Equip legendary creature {3}", "{3}",
				TargetCreature("target legendary creature you control", YouControl(), Legendary())),
			EquipAbility("{7}"),
		},
	})
}
