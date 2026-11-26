package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Swift Warkite — Creature — Dragon {4}{B}{R}, 4/4 (Dragons of Tarkir, #5551):
//
//	"Flying
//	 When this creature enters, you may put a creature card with mana value 3 or less from your hand or graveyard onto the battlefield. That creature gains haste. Return it to your hand at the beginning of the next end step."
//
// Bonny Pall's pick from hand or graveyard, with Nahiri, the Harbinger's
// haste and end-step return once the card has really entered (the keyed tail
// is told what entered). A card that never arrives is not given haste and
// not returned. The haste has no end, as printed.
//
// No simplification.

func init() {
	effects.Register(effects.Spec{
		OracleID:        "0df3444c-ae29-4fcc-bf46-4edf52b90e30",
		Name:            "Swift Warkite",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(swiftWarkiteLabel, swiftWarkiteOffer),
		},
	})
}

const swiftWarkiteLabel = "Swift Warkite — you may put a creature card with mana value 3 or less from your hand or graveyard onto the battlefield; it gains haste; return it to your hand at the beginning of the next end step"

// swiftWarkiteOffer asks for the creature card: a nontoken creature card of
// mana value 3 or less in the controller's hand or graveyard.
func swiftWarkiteOffer(g *game.Game, item *game.StackItem) error {
	p := g.PlayerByIDForEffect(item.Controller)
	if p == nil {
		return nil
	}
	var ids []uuid.UUID
	for _, z := range []*game.Zone{p.Hand, p.Graveyard} {
		if z == nil {
			continue
		}
		for _, c := range z.Cards {
			if c.IsCreature() && !c.IsToken() && c.ManaValue() <= 3 {
				ids = append(ids, c.InstanceID)
			}
		}
	}
	if len(ids) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: swiftWarkiteLabel,
		Cards:    ids,
		Min:      0,
		Max:      1,
		ThenKey:  swiftWarkitePicked,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// swiftWarkitePicked puts the picked card onto the battlefield from the zone
// that holds it now. An on-disk key: never renamed, never reused.
var swiftWarkitePicked = game.RegisterChooseCardsThen("choose-cards/swift-warkite-put", func(g *game.Game, r game.CardsChosen) error {
	if !r.Answered || len(r.Picked) == 0 || r.Asked == nil {
		return nil
	}
	id := r.Picked[0]
	z := g.FindCardZoneForEffect(id)
	if z == nil {
		return nil
	}
	tail := effects.NewContext(g, r.Asked).ItemTail(swiftWarkiteEntered)
	switch z.Kind {
	case game.ZoneHand:
		return g.PutFromHandOntoBattlefieldKeyedForEffect(id, game.HandEntryOptions{Controller: r.Chooser}, tail)
	case game.ZoneGraveyard:
		return g.ReturnFromGraveyardToBattlefieldKeyedForEffect(id, r.Chooser, false, tail)
	}
	return nil
})

// swiftWarkiteEntered is "That creature gains haste. Return it to your hand
// at the beginning of the next end step." once the creature has entered.
var swiftWarkiteEntered = effects.EnteredTail("swift-warkite/haste-and-return", func(ctx *effects.Context, entered uuid.UUID, _ game.TailOutcome) error {
	if entered == uuid.Nil {
		return nil
	}
	c, ok := effects.BattlefieldCardForEffect(ctx.Game, entered)
	if !ok {
		return nil
	}
	if err := (effects.ScopedEffectFor{
		Target:   entered,
		Mods:     []game.Mod{game.AddKeywordsMod("haste")},
		Duration: game.IndefiniteDuration(),
		Label:    "Swift Warkite — haste",
	}).Apply(ctx); err != nil {
		return err
	}
	return effects.ScheduleDelayedTrigger{
		At:     game.StepEnd,
		Label:  "Swift Warkite — return " + c.Name + " to your hand",
		Cards:  []uuid.UUID{entered},
		Body:   effects.ZaraReturnToHandBody,
		Params: game.EffectParams{Object: game.ObjectRef{ID: entered, Epoch: c.ObjectEpoch}},
	}.Apply(ctx)
})
