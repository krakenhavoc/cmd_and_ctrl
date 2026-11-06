package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Creeping Crystal Coating — Enchantment — Aura {2}{G} (Avatar: The Last
// Airbender Eternal, #5075):
//
//	"Flash
//	 Enchant creature
//	 Enchanted creature gets +0/+3 and has "Whenever this creature
//	 attacks, create a Food token.""
//
// Bear Umbra's shape: a layer 7c pump (CR 613.4c) and a granted attack
// trigger on the host. The granted trigger is the host's own, so "this
// creature" is the enchanted creature, and its Food is made for the
// host's controller. It is written with Do, so the trigger carries a
// footprint.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c23f2546-bf9e-4390-b6f7-c616920e04ca",
		Name:            "Creeping Crystal Coating",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         EnchantCreature(),
		Purpose:         game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			PumpAttached(0, 3),
			GrantAbilitiesToAttached(creepingCrystalCoatingGrant),
		},
		Grants: []AbilityGrant{{
			Key: creepingCrystalCoatingGrant,
			Triggered: []game.TriggeredAbility{
				WheneverThisAttacks("Creeping Crystal Coating — create a Food token", Do(CreateToken{Template: FoodToken(), N: 1})),
			},
			Text: "Whenever this creature attacks, create a Food token.",
		}},
	})
}

const creepingCrystalCoatingGrant = "avatar-b30/creeping-crystal-coating-food"
