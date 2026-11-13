package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sleeper's Robe — Enchantment — Aura {U}{B} (Invasion, #5616):
//
//	"Enchant creature
//	 Enchanted creature has fear.
//	 Whenever enchanted creature deals combat damage to an opponent,
//	 you may draw a card."
//
// Fear is a layer-6 grant (GrantToAttached; CR 702.36b). The draw is
// Curiosity's optional draw with a narrower condition: COMBAT damage
// only (attachedCreatureDealtCombatDamageToPlayer), and only to an
// opponent of the Aura's controller (CR 109.5), whoever controls the
// creature. "You may" is a real prompt to the Aura's controller.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d82783d9-9bfe-4fcd-a7d2-6e2477f99edc",
		Name:         "Sleeper's Robe",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			GrantToAttached("fear"),
		},
		Triggered: []game.TriggeredAbility{
			OptionalDraw(On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Target != source.Controller && attachedCreatureDealtCombatDamageToPlayer(ev, source, g)
			}, "Sleeper's Robe — draw a card", Do(DrawCards{N: 1})), "Sleeper's Robe — draw a card?", 1),
		},
	})
}
