package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vampiric Link — Enchantment — Aura {B} (Planar Chaos, #5595):
//
//	"Enchant creature
//	 Whenever enchanted creature deals damage, you gain that much life."
//
// Spirit Link in black: any damage the enchanted creature deals, combat or
// not, to any player or permanent, gains the Aura's controller that much
// life, read off the event after any prevention or reduction.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3f4521b4-1efd-4592-a50d-5ee9918df177",
		Name:         "Vampiric Link",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Amount > 0 && source.IsAttachedTo(ev.Source)
			},
			Key: "Vampiric Link — gain that much life",
			Effect: func(g *game.Game, item *game.StackItem) error {
				return GainLife{Player: item.Controller, Amount: item.Trigger.Event.Amount}.Apply(NewContext(g, item))
			},
		}},
	})
}
