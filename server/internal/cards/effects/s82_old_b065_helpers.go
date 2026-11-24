package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// EnchantedPermanentLeft is "when enchanted creature leaves the
// battlefield": any departure of the permanent this Aura is attached to
// (Valor of the Worthy). The event is harvested while the Aura is still
// attached, before the CR 704.5m unattach, so the attachment names it.
func EnchantedPermanentLeft(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventLTB && source.IsAttachedTo(ev.CardID)
}

// TapEnchantedAndItDamagesYou is Bind the Monster's entry trigger: tap
// the creature the Aura enchants, which then deals damage equal to its
// power to the Aura's controller.
func TapEnchantedAndItDamagesYou(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	aura, ok := ctx.TriggeringPermanent()
	if !ok || aura.AttachedTo.Kind != game.TargetCard {
		return nil
	}
	host := aura.AttachedTo.ID
	if err := (TapTarget{Target: host}).Apply(ctx); err != nil {
		return err
	}
	c, ok := g.LookupCardForEffect(host)
	if !ok {
		return nil
	}
	power := c.CurrentPower()
	if power <= 0 {
		return nil
	}
	return g.DealDamageToPlayerForEffect(host, item.Controller, power)
}
