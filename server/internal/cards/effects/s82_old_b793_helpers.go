package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// OldB793DealsOneToTargetAndTo is "<this> deals 1 damage to any target and 1
// damage to <someone>" (Granger Guildmage: you; Reckless Embermage: itself):
// one source, one simultaneous damage event (CR 120.2). A target that is no
// longer legal makes the whole ability do nothing (CR 608.2b), the second
// damage included, since both halves belong to the one resolving ability.
// `other` names the second recipient as the ability resolves.
func OldB793DealsOneToTargetAndTo(other func(ctx *Context) uuid.UUID) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		legal := ctx.LegalTargets()
		if len(legal) == 0 {
			return nil
		}
		return g.DamageInstanceForEffect(func() error {
			for _, t := range legal {
				if err := (DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: 1}).Apply(ctx); err != nil {
					return err
				}
			}
			return DealDamage{Source: item.SourceCardID, Target: other(ctx), Amount: 1}.Apply(ctx)
		})
	}
}

// OldB793ThisDealtDamageGainThatMuch is "Whenever this creature deals damage,
// you gain that much life": damage of any kind, to anything.
func OldB793ThisDealtDamageGainThatMuch(label string) game.TriggeredAbility {
	return On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
		return ev.Kind == game.EventDealDamage && ev.Amount > 0 && ev.Source == source.InstanceID
	}, label, func(g *game.Game, item *game.StackItem) error {
		return GainLife{Player: item.Controller, Amount: item.Trigger.Event.Amount}.Apply(NewContext(g, item))
	})
}
