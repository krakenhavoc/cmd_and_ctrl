package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flight of Fancy — Enchantment — Aura {3}{U} (Ravnica: City of Guilds, S82 OLD
// batch 581, #5600):
//
//	"Enchant creature
//	 When this Aura enters, draw two cards.
//	 Enchanted creature has flying."
//
// The draw is the Aura's own enter trigger (the controller of the Aura draws,
// whoever controls the creature); the flying is a layer-6 grant to the
// attached creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cd20a2a4-5e5e-420d-9420-651bce511f76",
		Name:         "Flight of Fancy",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit, Draws: 2},
		Static: []game.StaticAbility{
			effects.GrantToAttached("flying"),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Flight of Fancy — draw two cards", effects.Do(effects.DrawCards{N: 2})),
		},
	})
}
