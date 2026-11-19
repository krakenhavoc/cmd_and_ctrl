package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "a621c5f6-c1c6-4bc9-88b1-f459efa07a39",
		Name:         "Smoke Shroud",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
			GrantToAttached("flying"),
		},
		Triggered: []game.TriggeredAbility{
			func() game.TriggeredAbility {
				t := Optional(On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					c, ok := enteredUnderYourControl(ev, source, g, false)
					return ok && c.IsCreature() && c.HasSubtypeWord("Ninja")
				}, "Smoke Shroud — you may return this card from your graveyard to the battlefield attached to that creature",
					smokeShroudReturn), "Smoke Shroud — return it attached to that Ninja?")
				t.Zones = []game.ZoneKind{game.ZoneGraveyard}
				return t
			}(),
		},
	})
}
