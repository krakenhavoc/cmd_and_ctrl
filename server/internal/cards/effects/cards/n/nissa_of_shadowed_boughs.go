package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nissa of Shadowed Boughs — Legendary Planeswalker — Nissa {2}{B}{G}, loyalty
// 4 (Zendikar Rising, #5508):
//
//	"Landfall — Whenever a land you control enters, put a loyalty counter on
//	 Nissa.
//	 +1: Untap target land you control. You may have it become a 3/3
//	 Elemental creature with haste and menace until end of turn. It's still a
//	 land.
//	 −5: You may put a creature card with mana value less than or equal to the
//	 number of lands you control onto the battlefield from your hand or
//	 graveyard with two +1/+1 counters on it."
//
// The landfall trigger is Landfall + CounterOnThis with the loyalty kind. The
// +1 untaps the target, then asks "you may" (CR 608.2d): on yes, a timed
// effect on the land (layers 4, 7b and 6) makes it a 3/3 Elemental creature
// with haste and menace until end of turn, still a land. The −5 picks, from the
// controller's hand and graveyard, a creature card whose mana value is at most
// the lands they control as it resolves, and puts it onto the battlefield with
// two +1/+1 counters (they enter with them, CR 122.6); declining is allowed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "15e3e226-0dd1-4c9b-a7ec-af67509cbfc5",
		Name:            "Nissa of Shadowed Boughs",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 4,
		Triggered: []game.TriggeredAbility{
			effects.Landfall("Nissa of Shadowed Boughs — landfall: put a loyalty counter on Nissa",
				effects.Do(effects.CounterOnThis{Kind: game.CounterLoyalty, N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "+1: Untap target land you control. You may have it become a 3/3 Elemental creature with haste and menace until end of turn. It's still a land.",
			Cost:    effects.LoyaltyCost(1),
			Targets: effects.TargetPermanent("target land you control", effects.Land(), effects.YouControl()),
			Effect:  nissaShadowedUntapAndAnimate,
		}, {
			Label:  "−5: You may put a creature card with mana value less than or equal to the number of lands you control onto the battlefield from your hand or graveyard with two +1/+1 counters on it.",
			Cost:   effects.LoyaltyCost(-5),
			Effect: nissaShadowedReanimate,
		}},
	})
}

func nissaShadowedUntapAndAnimate(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := (effects.UntapTarget{Target: id}).Apply(ctx); err != nil {
		return err
	}
	return effects.MayChoice{
		Question: "Nissa of Shadowed Boughs — have it become a 3/3 Elemental creature with haste and menace until end of turn?",
		Then:     nissaShadowedAnimate,
		Carry:    []uuid.UUID{id},
	}.Apply(ctx)
}

// nissaShadowedAnimate is the "yes" to the animation: the untapped land becomes a
// 3/3 Elemental creature with haste and menace until end of turn. An on-disk key.
var nissaShadowedAnimate = effects.MayYes("confirm/nissa-of-shadowed-boughs-animate", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	if len(r.Carry) == 0 || !ctx.Game.Battlefield.Contains(r.Carry[0]) {
		return nil
	}
	return effects.ScopedEffectFor{
		Target: r.Carry[0],
		Mods: []game.Mod{
			game.AddTypesMod("Creature"),
			game.AddSubtypesMod(ctx.TextSubtypes("Elemental")...),
			game.SetBasePowerMod(3),
			game.SetBaseToughnessMod(3),
			game.AddKeywordsMod("haste", "menace"),
		},
		Duration: effects.DurationUntilEndOfTurn(ctx),
		Label:    "Nissa of Shadowed Boughs — becomes a 3/3 Elemental creature with haste and menace until end of turn",
	}.Apply(ctx)
})

func nissaShadowedReanimate(g *game.Game, item *game.StackItem) error {
	you := item.Controller
	p := g.PlayerByIDForEffect(you)
	if p == nil {
		return nil
	}
	lands := effects.CountControlled(g, you, func(c game.Card) bool { return c.IsLand() })
	var ids []uuid.UUID
	for _, z := range []*game.Zone{p.Hand, p.Graveyard} {
		if z == nil {
			continue
		}
		for _, c := range z.Cards {
			if c.IsCreature() && !c.IsToken() && c.ManaValue() <= lands {
				ids = append(ids, c.InstanceID)
			}
		}
	}
	if len(ids) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  you,
		Source:   item.SourceCardID,
		Question: "Nissa of Shadowed Boughs — you may put a creature card with mana value at most the number of lands you control onto the battlefield from your hand or graveyard with two +1/+1 counters on it",
		Cards:    ids,
		Min:      0,
		Max:      1,
		ThenKey:  nissaShadowedPut,
	})
	return nil
}

// nissaShadowedPut is the chosen creature card entering from its zone with
// two +1/+1 counters. An on-disk key.
var nissaShadowedPut = game.RegisterChooseCardsThen("choose-cards/nissa-of-shadowed-boughs-put", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	id := r.Picked[0]
	z := g.FindCardZoneForEffect(id)
	if z == nil {
		return nil
	}
	counters := map[string]int{game.CounterPlusOne: 2}
	switch z.Kind {
	case game.ZoneHand:
		return g.PutFromHandOntoBattlefieldKeyedForEffect(id, game.HandEntryOptions{Controller: r.Chooser, EntersWithCounters: counters}, game.KeyedTail{})
	case game.ZoneGraveyard:
		return g.ReturnFromGraveyardWithCountersKeyedForEffect(id, r.Chooser, false, counters, game.KeyedTail{})
	}
	return nil
})
