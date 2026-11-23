package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Smoke Shroud — Enchantment — Aura {1}{U} (Foundations Jumpstart):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+1 and has flying.
//	 When a Ninja you control enters, you may return this card from your
//	 graveyard to the battlefield attached to that creature."
//
// The trigger watches from the graveyard only (CR 113.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a621c5f6-c1c6-4bc9-88b1-f459efa07a39",
		Name:         "Smoke Shroud",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 1),
			effects.GrantToAttached("flying"),
		},
		Triggered: []game.TriggeredAbility{
			func() game.TriggeredAbility {
				t := effects.Optional(effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
					return ok && c.IsCreature() && c.HasSubtypeWord("Ninja")
				}, "Smoke Shroud — you may return this card from your graveyard to the battlefield attached to that creature",
					effects.SmokeShroudReturn), "Smoke Shroud — return it attached to that Ninja?")
				t.Zones = []game.ZoneKind{game.ZoneGraveyard}
				return t
			}(),
		},
	})
}
