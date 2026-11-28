package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Frog Tongue — Enchantment — Aura {G} (Tempest, S82 OLD, #5628):
//
//	"Enchant creature
//	 When this Aura enters, draw a card.
//	 Enchanted creature has reach. (It can block creatures with flying.)"
//
// An enters trigger that draws for the Aura's controller, and a layer 6 keyword
// grant to whatever it is attached to.
//
// No simplification.
func init() {
	const name = "Frog Tongue"
	effects.Register(effects.Spec{
		OracleID:     "861c5374-a8e7-4684-9a1b-e6ce3b98e4a2",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(name+" — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
		Static: []game.StaticAbility{effects.GrantToAttached("reach")},
	})
}
