package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Leaden Fists — Enchantment — Aura {2}{U} (Future Sight, #5594):
//
//	"Flash (You may cast this spell any time you could cast an instant.)
//	 Enchant creature
//	 Enchanted creature gets +3/+3 and doesn't untap during its
//	 controller's untap step."
//
// The shared attached pump with Charmed Sleep's untap-step restriction
// (CR 502.3). Nothing taps the creature: an untapped one keeps attacking
// or blocking until it is tapped for the first time.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:              "d4bcbce9-9739-4828-9980-e21490b44346",
		Name:                  "Leaden Fists",
		Completeness:          CompletenessFull,
		PrintedKeywords:       []string{"flash"},
		Targets:               EnchantCreature(),
		Purpose:               game.Purpose{Aura: game.AuraBenefit},
		Static:                []game.StaticAbility{PumpAttached(3, 3)},
		UntapStepRestrictions: []game.UntapStepRestriction{enchantedDoesntUntap()},
	})
}
