package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Future Flight — {2}{U}{U} Enchantment — Aura (Marvel's Spider-Man Eternal,
// #4460):
//
//	"Enchant creature
//	 When this Aura enters, draw two cards.
//	 Enchanted creature gets +2/+0 and has flying."
//
// Aether Tunnel's pump and keyword grant plus an entry draw.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bfd5b48a-698e-4ade-afb0-954ba288f9a7",
		Name:         "Future Flight",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 0),
			effects.GrantToAttached("flying"),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Future Flight — draw two cards", effects.Do(effects.DrawCards{N: 2})),
		},
	})
}
