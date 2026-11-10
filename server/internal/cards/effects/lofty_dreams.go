package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lofty Dreams — Enchantment — Aura {3}{U}{U} (Lorwyn Eclipsed, #4458):
//
//	"Convoke (Your creatures can help cast this spell. Each creature you
//	 tap while casting this spell pays for {1} or one mana of that
//	 creature's color.)
//	 Enchant creature
//	 When this Aura enters, draw a card.
//	 Enchanted creature gets +2/+2 and has flying."
//
// Feather of Flight's Aura (an entry draw that uses the stack, the shared
// attachment pump and flying grant) with Appeal to Eirdu's convoke. The
// creature convoked with may be the one enchanted.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f75568a0-33ed-45a0-a885-48a19a5bb58b",
		Name:         "Lofty Dreams",
		Completeness: CompletenessFull,
		TapCost:      Convoke(),
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit, Draws: 1},
		Static: []game.StaticAbility{
			PumpAttached(2, 2),
			GrantToAttached("flying"),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Lofty Dreams — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
