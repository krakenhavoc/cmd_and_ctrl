package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Armadillo Cloak — Enchantment — Aura {1}{G}{W} (Invasion, #5616):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2 and has trample.
//	 Whenever enchanted creature deals damage, you gain that much life."
//
// Frantic Strength's two statics (PumpAttached, GrantToAttached) and
// Spirit Link's damage trigger: any damage the enchanted creature deals,
// combat or not, to a player or to a creature, read off the event's
// Amount (the damage actually dealt, after prevention). The life goes to
// the Aura's controller (CR 109.5's "you"), whoever controls the
// creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "15397a59-70a8-441f-abf7-230405a1602b",
		Name:         "Armadillo Cloak",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			PumpAttached(2, 2),
			GrantToAttached("trample"),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Amount > 0 && source.IsAttachedTo(ev.Source)
			},
			Key: "Armadillo Cloak — gain that much life",
			Effect: func(g *game.Game, item *game.StackItem) error {
				return GainLife{Player: item.Controller, Amount: item.Trigger.Event.Amount}.Apply(NewContext(g, item))
			},
		}},
	})
}
