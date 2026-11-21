package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Strong Back — Enchantment — Aura {2}{G} (Fallout):
//
//	"Enchant creature
//	 Equip abilities you activate that target enchanted creature cost {3}
//	 less to activate.
//	 Aura spells you cast that target enchanted creature cost {3} less to
//	 cast.
//	 Enchanted creature gets +2/+2 for each Aura and Equipment attached to
//	 it."
//
// Two cost modifiers read the announced targets (CR 601.2c comes before
// 601.2f, Killian's reading) and ask whether one of them is the creature
// Strong Back is attached to (gapTargetsEnchantedCreature); each takes {3}
// of generic mana off, never coloured mana. The pump is With Great Power's
// count of the Auras and Equipment on the host, this Aura included.
//
// No simplification.
func init() {
	equip := ActivationCostsLess(3, "Equip abilities you activate that target enchanted creature cost {3} less to activate.",
		AnEquipAbilityCost(), YouActivate(), gapTargetsEnchantedCreature())
	equip.ReadsTargets = true
	aura := CostsLess(3, "Aura spells you cast that target enchanted creature cost {3} less to cast.",
		YourSpell(), gapAuraSpell(), gapTargetsEnchantedCreature())
	aura.ReadsTargets = true
	Register(Spec{
		OracleID:      "097992a9-7b88-4c3f-8191-1ff10a4a735d",
		Name:          "Strong Back",
		Completeness:  CompletenessFull,
		Targets:       EnchantCreature(),
		CostModifiers: []game.CostModifier{equip, aura},
		Static:        []game.StaticAbility{PumpAttachedPer(2, 2, aurasAndEquipmentOnTheHost)},
	})
}
