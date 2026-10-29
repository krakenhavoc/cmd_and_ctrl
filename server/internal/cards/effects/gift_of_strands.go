package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gift of Strands — Enchantment — Aura {3}{G} (The Hobbit Eternal, HOC):
//
//	"Flash
//	 Enchant creature
//	 When this Aura enters, scry 2.
//	 Enchanted creature gets +3/+3."
//
// Feather of Flight's shape: flash rides PrintedKeywords, the Aura
// targets a creature as it is cast, the pump is PumpAttached, and the
// scry is a trigger on the Aura's own entry (it uses the stack and can be
// responded to). If the creature is gone as the Aura resolves, the Aura
// is not put onto the battlefield and never scries (CR 303.4g, #4454).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7a7789e5-f2ac-49b2-943e-7d49e4727fb6",
		Name:            "Gift of Strands",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         EnchantCreature(),
		Purpose:         game.Purpose{Aura: game.AuraBenefit},
		Static:          []game.StaticAbility{PumpAttached(3, 3)},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Gift of Strands — scry 2", Do(Scry{N: 2})),
		},
	})
}
