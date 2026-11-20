package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wolfkin Bond — Enchantment — Aura {4}{G} (Eldritch Moon, #5541):
//
//	"Enchant creature
//	 When this Aura enters, create a 2/2 green Wolf creature token.
//	 Enchanted creature gets +2/+2."
//
// The Wolf is the token table's row; the pump is +2 Mace's on the host.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "49e22e88-d4d9-4781-81e3-bf77b76b132b",
		Name:         "Wolfkin Bond",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{PumpAttached(2, 2)},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Wolfkin Bond — create a 2/2 green Wolf creature token",
				Do(CreateToken{Template: TokenCard("2/2 green Wolf"), N: 1})),
		},
	})
}
