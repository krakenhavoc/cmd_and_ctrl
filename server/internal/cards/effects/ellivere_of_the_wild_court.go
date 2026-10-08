package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ellivere of the Wild Court — Legendary Creature — Human Knight
// {2}{G}{W}, 4/4:
//
//	"Whenever Ellivere enters or attacks, create a Virtuous Role token
//	 attached to another target creature you control. (If you control
//	 another Role on it, put that one into the graveyard. Enchanted
//	 creature gets +1/+1 for each enchantment you control.)
//	 Whenever an enchanted creature you control deals combat damage to a
//	 player, draw a card."
//
// The first line is ONE ability with two trigger conditions. "Another"
// excludes Ellivere by object, so the target is required (and the
// trigger is removed with no other creature, CR 603.3d). The second is
// one trigger per creature that connects, not per combat: an enchanted
// creature is any creature with an Aura attached, the Role included.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "97f9b82f-b3cf-44ff-9abc-2d92d7cbaa27",
		Name:         "Ellivere of the Wild Court",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEntersOrAttacks("Ellivere of the Wild Court — create a Virtuous Role token attached to another target creature you control",
					createRoleOnFirstTarget(RoleVirtuous)),
				Another(TargetCreature("another target creature you control", YouControl()))),
			On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return combatDamageToPlayerBy(ev, source.Controller, g) && isEnchantedByAura(g, ev.Source)
			}, "Ellivere of the Wild Court — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
