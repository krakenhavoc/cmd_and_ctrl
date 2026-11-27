package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b578_helpers.go — shared pieces of Ravnica: City of Guilds batch 578
// (#5600).

// ThisDealsCombatDamageYouGainThatMuchLife is "Whenever this creature deals
// combat damage, you gain that much life" (Sunhome Enforcer, and the ability
// Flash Conscription grants): one trigger per combat damage event the creature
// is the source of, to a player, a planeswalker or a creature. The amount is the
// damage the event carried, read as the ability resolves; the life is gained by
// the ability's controller, which for a stolen creature is the thief.
func ThisDealsCombatDamageYouGainThatMuchLife(name string) game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches: []game.EventKind{game.EventDealDamage},
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
			return ev.Combat && ev.Amount > 0 && ev.Source == source.InstanceID
		},
		Key: name + " — you gain that much life",
		Effect: func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			return GainLife{Player: ctx.Controller(), Amount: item.Trigger.Event.Amount}.Apply(ctx)
		},
	}
}
