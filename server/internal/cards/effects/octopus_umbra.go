package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Octopus Umbra — Enchantment — Aura {3}{U}{U} (CR 702.89, #3654):
//
//	"Enchant creature
//	 Enchanted creature has base power and toughness 8/8 and has "Whenever
//	 this creature attacks, you may tap target creature with power 8 or
//	 less."
//	 Umbra armor (If enchanted creature would be destroyed, instead remove
//	 all damage from it and destroy this Aura.)"
//
// The base P/T is a layer 7b set (SetAttachedBasePT). The quoted attack
// trigger is a granted bundle (ADR 0093) with a targeted, optional
// effect, shaped as Clever Distraction's. No simplification.
func init() {
	Register(Spec{
		OracleID:     octopusUmbraOracleID,
		Name:         "Octopus Umbra",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []AbilityGrant{{
			Key: octopusUmbraTap,
			Triggered: []game.TriggeredAbility{
				Optional(game.TriggeredAbility{
					Watches:   []game.EventKind{game.EventAttack},
					AppliesTo: ThisAttacked,
					Targets:   TargetCreature("target creature with power 8 or less", PowerLE(8)),
					Key:       "Octopus Umbra — tap target creature with power 8 or less",
					Effect:    b36TapChosenCreature,
				}, "Octopus Umbra — tap target creature?"),
			},
			Text: "Whenever this creature attacks, you may tap target creature with power 8 or less.",
		}},
		Static: []game.StaticAbility{
			SetAttachedBasePT(8, 8),
			GrantAbilitiesToAttached(octopusUmbraTap),
		},
		Replacements: []game.ReplacementEffect{UmbraArmor()},
	})
}

const (
	octopusUmbraOracleID = "6d2ce0ba-f2ff-4c2b-ba85-f6c011ebda6c"
	octopusUmbraTap      = "octopus-umbra/tap"
)
