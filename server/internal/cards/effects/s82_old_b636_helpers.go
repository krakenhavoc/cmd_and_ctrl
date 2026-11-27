package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b636_helpers.go — Planeshift Auras, batch 636 (#5615).

// S82OldB636InsolenceDamage is "this Aura deals 2 damage to that creature's
// controller": the controller of the enchanted creature as it became tapped,
// read now (CR 608.2h), is dealt 2 damage by the Aura.
func S82OldB636InsolenceDamage(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil || item.Trigger.Event.CardID == uuid.Nil {
		return nil
	}
	creature, ok := g.LookupCardForEffect(item.Trigger.Event.CardID)
	if !ok {
		return nil
	}
	return DealDamage{Source: item.SourceCardID, Target: creature.Controller, Amount: 2}.Apply(NewContext(g, item))
}

// S82OldB636NotSharingAMostCommonColor is Heroic Defiance's condition: the
// enchanted creature does not share a colour with the most common colour among
// all permanents, or a colour tied for most common. It is read afresh on every
// recompute (layer 5 has settled by layer 7c, CR 613.1).
func S82OldB636NotSharingAMostCommonColor(target *game.Card, g *game.Game, source *game.Card) bool {
	return AttachedToSource(target, g, source) && !SharesAMostCommonColor(g, target)
}

// S82OldB636TargetBecomesAColorOfYourChoice is the body of "Target creature
// becomes the color of your choice until end of turn" as an ability's whole
// effect (Sisay's Ingenuity's granted ability): the controller of the ability
// names the colour as it resolves, and a target that is gone does nothing.
func S82OldB636TargetBecomesAColorOfYourChoice(name string, then game.TailRef) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		if len(ctx.LegalTargets()) == 0 {
			return nil
		}
		ChooseColorTail(game.ColorForBenefit, g, ctx.Controller(), ctx.Source(),
			name+" — choose a color", ctx.ItemTail(then))
		return nil
	}
}

// S82OldB636SacrificeThisAura is "sacrifice this Aura" as a trigger's whole
// effect (Sleeping Potion): it does nothing once the Aura has left the
// battlefield (CR 608.2b).
func S82OldB636SacrificeThisAura(g *game.Game, item *game.StackItem) error {
	if !OnBattlefield(g, item.SourceCardID) {
		return nil
	}
	return SacrificePermanent{Target: item.SourceCardID}.Apply(NewContext(g, item))
}
