package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_gap_b139_helpers.go — Dominaria Remastered white creatures and
// enchantments, GAP batch 139 (#5948).

// GapB139WheneverThisDealsDamageYouGainThatMuchLife is "Whenever this
// creature deals damage, you gain that much life" (Kjeldoran Gargoyle): any
// damage, combat or not, to anything. The amount is the damage the event
// carried, read when the ability resolves; the life is the controller's.
func GapB139WheneverThisDealsDamageYouGainThatMuchLife(name string) game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches: []game.EventKind{game.EventDealDamage},
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
			return ev.Amount > 0 && ev.Source == source.InstanceID
		},
		Key: name + " — you gain that much life",
		Effect: func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			return GainLife{Player: ctx.Controller(), Amount: item.Trigger.Event.Amount}.Apply(ctx)
		},
	}
}
