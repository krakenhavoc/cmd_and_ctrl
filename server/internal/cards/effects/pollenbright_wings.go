package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pollenbright Wings — Enchantment — Aura {4}{G}{W} (Wilds of Eldraine
// Commander, #5368):
//
//	"Enchant creature
//	 Enchanted creature has flying.
//	 Whenever enchanted creature deals combat damage to a player, create
//	 that many 1/1 green Saproling creature tokens."
//
// The trigger is the Aura's own (its controller gets the tokens, not the
// creature's), read off the attachment as Celestial Mantle's is; "that
// many" is the damage dealt, read off the event as Lathril's is. Combat
// damage to a planeswalker or battle does not count.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "582c8030-1d41-4c63-8ed6-06aad398a96c",
		Name:         "Pollenbright Wings",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{GrantToAttached("flying")},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attachedCreatureDealtCombatDamageToPlayer(ev, source, g)
			}, "Pollenbright Wings — create that many 1/1 green Saproling creature tokens",
				Do(CreateTokensForEach{
					Template: TokenCard("1/1 green Saproling"),
					Name:     "the combat damage dealt",
					Count:    func(ctx *Context) int { return ctx.Trigger().Event.Amount },
				})),
		},
	})
}
