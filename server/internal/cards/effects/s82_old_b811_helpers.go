package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b811_helpers.go — Mirage Auras (#5633).

// S82OldB811ConsumingFerocityUpkeep is Consuming Ferocity's upkeep trigger: a
// +1/+0 counter goes on the enchanted creature, and if it then has three or
// more +1/+0 counters it deals damage equal to its power to its controller,
// then it is destroyed and can't be regenerated. The creature is the one the
// Aura is attached to as the trigger resolves; the damage is read from the
// creature as it is (CR 608.2h), and the destruction happens only to the same
// object (CR 400.7).
func S82OldB811ConsumingFerocityUpkeep(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	host := EnchantedCreatureOf(ctx)
	if host == uuid.Nil || !OnBattlefield(g, host) {
		return nil
	}
	if err := (AddCounter{Target: host, Kind: "+1/+0", N: 1}).Apply(ctx); err != nil {
		return err
	}
	c, ok := g.LookupCardForEffect(host)
	if !ok || c.Counters["+1/+0"] < 3 {
		return nil
	}
	if err := (DealDamage{Source: host, Target: c.Controller, Amount: B31CurrentPowerOf(g, host)}).Apply(ctx); err != nil {
		return err
	}
	if !OnBattlefield(g, host) {
		return nil
	}
	return DestroyTarget{Target: host, CantBeRegenerated: true}.Apply(ctx)
}

// S82OldB811ControllerHasAnotherCreature is "its controller controls another
// creature", for the host of an Aura: any creature under the host's
// controller's control but the host itself, in the layered view.
func S82OldB811ControllerHasAnotherCreature(host *game.Card, g *game.Game, _ *game.Card) bool {
	return CountControlled(g, host.Controller, func(c game.Card) bool {
		return c.InstanceID != host.InstanceID && c.IsCreature()
	}) > 0
}

// S82OldB811ReflectDamageToEnchantedCreaturesController is
// Binding Agony's trigger: "Whenever enchanted creature is dealt damage, this
// Aura deals that much damage to that creature's controller." The creature's
// controller and the amount are recorded as the trigger goes on the stack, so
// a creature that dies to the damage still costs its controller.
func S82OldB811ReflectDamageToEnchantedCreaturesController(label string) game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches: []game.EventKind{game.EventDealDamage},
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
			return ev.Amount > 0 && source.AttachedTo.Kind == game.TargetCard && source.AttachedTo.ID == ev.Target
		},
		Key: label,
		Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
			item := game.NewTriggeredItem(source, label)
			item.Params.Amount = ev.Amount
			if c, ok := g.LookupCardForEffect(ev.Target); ok {
				item.Params.Player = c.Controller
			}
			return item
		},
		Effect: S82OldB811DealRecordedDamageToRecordedPlayer,
	}
}

// S82OldB811DealRecordedDamageToRecordedPlayer is the resolution of "this deals that
// much damage to that player": the source deals the amount and the player the trigger
// recorded as it went on the stack (item.Params.Amount and item.Params.Player), so a
// creature that died to the damage still costs its controller (Binding Agony, Justice).
func S82OldB811DealRecordedDamageToRecordedPlayer(g *game.Game, item *game.StackItem) error {
	if item.Params.Player == uuid.Nil || item.Params.Amount <= 0 {
		return nil
	}
	return DealDamage{Source: item.SourceCardID, Target: item.Params.Player, Amount: item.Params.Amount}.Apply(NewContext(g, item))
}
