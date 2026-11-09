package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Firion, Wild Rose Warrior — Legendary Creature — Human Rebel Warrior
// {2}{R}, 3/3:
//
//	"Equipped creatures you control have haste.
//	 Whenever a nontoken Equipment you control enters, create a token
//	 that's a copy of it, except it has 'This Equipment's equip abilities
//	 cost {2} less to activate.' Sacrifice that token at the beginning of
//	 the next upkeep."
//
// The haste is a layer-6 grant over each creature of yours that has an
// Equipment attached, read on every recompute. The trigger copies the
// Equipment that entered (read off the entering event, last-known if it
// has gone, CR 608.2h); "except it has '…'" is a CR 707.9a ability grant
// on the token, whose cost change is the new AbilityGrant.CostModifiers
// (read off the token's composite key, so "this" is the token). The copy
// enters, then the sacrifice is a delayed trigger at the beginning of the
// next upkeep, whoever's it is (CR 603.7), listing the token.
func init() {
	Register(Spec{
		OracleID:     "df42c768-34ca-4ebf-8f66-eaccf7261f70",
		Name:         "Firion, Wild Rose Warrior",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key:  firionEquipCostGrant,
			Text: "This Equipment's equip abilities cost {2} less to activate.",
			CostModifiers: []game.CostModifier{
				ActivationCostsLess(2, "This Equipment's equip abilities cost {2} less to activate.",
					AnEquipAbilityCost(), thisPermanentsAbility()),
			},
		}},
		Static: []game.StaticAbility{
			KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.Controller == source.Controller && isEquippedCreature(g, *target)
			}, "haste"),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, firionNontokenEquipmentEntered,
				"Firion, Wild Rose Warrior — create a token copy of that Equipment, sacrificed at the beginning of the next upkeep",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil || item.Trigger.Event.CardID == uuid.Nil {
						return nil
					}
					ctx := NewContext(g, item)
					return CreateTokenCopiesTogether{
						Controller: item.Controller,
						Copies:     []uuid.UUID{item.Trigger.Event.CardID},
						Except:     func(t *game.Card) { TokenCopyGainsAbility(t, firionEquipCostGrant) },
						Then:       ctx.ItemTail(firionSacrificeTheCopy),
					}.Apply(ctx)
				}),
		},
	})
}

const firionEquipCostGrant = "firion-wild-rose-warrior/equip-cost-less"

// firionNontokenEquipmentEntered is the trigger condition: a nontoken
// Equipment entered under the source's controller's control.
func firionNontokenEquipmentEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, false)
	return ok && c.HasSubtype("Equipment") && !c.IsToken()
}

// firionSacrificeTheCopy is "Sacrifice that token at the beginning of the
// next upkeep", told the token that landed. An on-disk key.
var firionSacrificeTheCopy = CardTail("firion-wild-rose-warrior/sacrifice-the-copy", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return ScheduleDelayedTrigger{
		At:    game.StepUpkeep,
		Label: "Firion, Wild Rose Warrior — sacrifice the token",
		Cards: o.IDs,
		Body:  sacrificeListedCardsBody,
	}.Apply(ctx)
})
