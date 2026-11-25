package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Azor's Gateway // Sanctum of the Sun (Rivals of Ixalan, S82 OLD batch 218,
// #5531).
//
// Front face, Legendary Artifact {2}:
//
//	"{1}, {T}: Draw a card, then exile a card from your hand. If cards with five
//	 or more different mana values are exiled with Azor's Gateway, you gain 5
//	 life, untap Azor's Gateway, and transform it."
//
// Back face, Legendary Land:
//
//	"(Transforms from Azor's Gateway.)
//	 {T}: Add X mana of any one color, where X is your life total."
//
// The draw comes first, then you choose a card from your hand to exile (with an
// empty hand nothing is exiled), linked to this object (CR 607.2a). Then, once the
// card has landed in exile, the count of different mana values among the cards
// exiled with it is read (a land is mana value 0, a split card the sum of its
// halves): five or more gains 5 life, untaps the Gateway and transforms it, in
// that order. The land makes X mana of one colour you choose, X being your life
// total as the ability resolves (a life total of 0 or less makes none).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     azorsGatewayOracleID,
		Name:         "Azor's Gateway",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}: Draw a card, then exile a card from your hand. If cards with five or more different mana values are exiled with Azor's Gateway, you gain 5 life, untap Azor's Gateway, and transform it.",
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
			Effect:  azorsGatewayActivate,
		}},
	})
	effects.Register(effects.Spec{
		OracleID:     azorsGatewayOracleID + "#1",
		Name:         "Sanctum of the Sun",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost: effects.ManaAbilityCost{Tap: true},
			ProducedAmount: effects.ProducedOneColor(func(g *game.Game, controller, _ uuid.UUID) int {
				if p := g.PlayerByIDForEffect(controller); p != nil && p.Life > 0 {
					return p.Life
				}
				return 0
			}),
			Label: "{T}: Add X mana of any one color, where X is your life total",
		}},
	})
}

const azorsGatewayOracleID = "c0cbb347-b060-43ce-a9c5-8c835be3cf1b"

func azorsGatewayActivate(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if err := (effects.DrawCards{Player: item.Controller, N: 1}).Apply(ctx); err != nil {
		return err
	}
	var hand []uuid.UUID
	if p := g.PlayerByIDForEffect(item.Controller); p != nil && p.Hand != nil {
		for _, c := range p.Hand.Cards {
			hand = append(hand, c.InstanceID)
		}
	}
	if len(hand) == 0 {
		return azorsGatewayCheck(ctx)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   ctx.Source(),
		Question: "Azor's Gateway — exile a card from your hand",
		Cards:    hand,
		Min:      1,
		Max:      1,
		ThenKey:  azorsGatewayPicked,
		LiveZone: game.ZoneHand,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// azorsGatewayPicked exiles the chosen card, linked to the Gateway, then checks
// the mana values. An on-disk key: never renamed, never reused.
var azorsGatewayPicked = effects.ItemChoice("choose-cards/azors-gateway-exile", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return azorsGatewayCheck(ctx)
	}
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	return ctx.Game.ExileLinkedKeyedForEffect(ref, r.Picked[:1], ctx.ItemTail(azorsGatewayChecked))
})

// azorsGatewayChecked is the check once the card has landed in exile. An
// on-disk key: never renamed, never reused.
var azorsGatewayChecked = effects.CardTail("then/azors-gateway-check", func(ctx *effects.Context, _ game.TailOutcome) error {
	return azorsGatewayCheck(ctx)
})

// azorsGatewayCheck is "If cards with five or more different mana values are
// exiled with Azor's Gateway, you gain 5 life, untap Azor's Gateway, and
// transform it."
func azorsGatewayCheck(ctx *effects.Context) error {
	ref, ok := ctx.SourceRef()
	if !ok || effects.OldB218DistinctManaValues(ctx.Game, game.PermissionCardRef(ref)) < 5 {
		return nil
	}
	if err := (effects.GainLife{Player: ctx.Controller(), Amount: 5}).Apply(ctx); err != nil {
		return err
	}
	if err := (effects.UntapTarget{Target: ctx.Source()}).Apply(ctx); err != nil {
		return err
	}
	return effects.TransformThis{}.Apply(ctx)
}
