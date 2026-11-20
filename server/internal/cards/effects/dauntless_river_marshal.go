package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dauntless River Marshal — Creature — Human Soldier {1}{W}, 2/1 (Magic
// 2015, #5556):
//
//	"This creature gets +1/+1 as long as you control an Island.
//	 {3}{U}: Tap target creature."
//
// Coastal Bulwark's Island-gated pump (any permanent you control with the
// Island subtype counts) and Frostbridge Guard's tapper without the {T},
// so it can be activated the turn the Marshal arrives.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "544c4cb3-cd43-4670-9883-e9545dad9cd6",
		Name:         "Dauntless River Marshal",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "This creature gets +1/+1 as long as you control an Island.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: selfWhileYouControlA(Subtype("Island")),
			Apply:     pumpBy(1, 1),
		}},
		Activated: []ActivatedAbility{{
			Label:   "{3}{U}: Tap target creature.",
			Cost:    ManaCost("{3}{U}"),
			Targets: TargetCreature("target creature"),
			Effect:  tapChosenPermanent,
		}},
	})
}
