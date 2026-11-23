package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "9ba6cec6-97bf-4a83-9ffa-77e187d0945b",
		Name:         "Knightly Valor",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 2),
			effects.GrantToAttached("vigilance"),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Knightly Valor — create a 2/2 white Knight creature token with vigilance",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 white Knight with vigilance"), N: 1})),
		},
	})
}
