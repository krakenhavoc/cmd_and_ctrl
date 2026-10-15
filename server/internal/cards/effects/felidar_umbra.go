package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Felidar Umbra — Enchantment — Aura {1}{W} (CR 702.89, #3654):
//
//	"Enchant creature
//	 Enchanted creature has lifelink.
//	 {1}{W}: Attach this Aura to target creature you control.
//	 Umbra armor (If enchanted creature would be destroyed, instead remove
//	 all damage from it and destroy this Aura.)"
//
// The move is an ordinary attach ability on the Aura (AttachSourceToTarget,
// as Equipment's equip is), with the target clause "creature you control".
// No simplification.
func init() {
	Register(Spec{
		OracleID:     felidarUmbraOracleID,
		Name:         "Felidar Umbra",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{GrantToAttached("lifelink")},
		Activated: []ActivatedAbility{{
			Label:   "{1}{W}: Attach this Aura to target creature you control",
			Cost:    ManaCost("{1}{W}"),
			Targets: TargetCreature("target creature you control", YouControl()),
			Effect:  AttachSourceToTarget,
		}},
		Replacements: []game.ReplacementEffect{UmbraArmor()},
	})
}

const felidarUmbraOracleID = "278574f6-e083-461b-8bac-52ed51f0cf63"
