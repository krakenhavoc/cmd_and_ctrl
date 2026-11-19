package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Knightly Valor — Enchantment — Aura {4}{W} (Enchanting Tales):
//
//	"Enchant creature
//	 When this Aura enters, create a 2/2 white Knight creature token with
//	 vigilance.
//	 Enchanted creature gets +2/+2 and has vigilance."
//
// The enchant clause is the Aura spell's target; the pump and the keyword are
// layer-6 and layer-7c statics on the host; the token is an ordinary ETB
// trigger of the Aura.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9ba6cec6-97bf-4a83-9ffa-77e187d0945b",
		Name:         "Knightly Valor",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static: []game.StaticAbility{
			PumpAttached(2, 2),
			GrantToAttached("vigilance"),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Knightly Valor — create a 2/2 white Knight creature token with vigilance",
				Do(CreateToken{Template: TokenCard("2/2 white Knight with vigilance"), N: 1})),
		},
	})
}
