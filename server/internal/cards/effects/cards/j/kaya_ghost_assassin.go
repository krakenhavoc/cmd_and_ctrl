package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kaya, Ghost Assassin — Legendary Planeswalker — Kaya {2}{W}{B}, loyalty 5:
//
//	"0: Exile Kaya or up to one target creature. Return that card to the
//	 battlefield under its owner's control at the beginning of your next
//	 upkeep. You lose 2 life.
//	 −1: Each opponent loses 2 life and you gain 2 life.
//	 −2: Each opponent discards a card and you draw a card."
//
// The 0 names "up to one target creature". With no target it exiles Kaya; with
// a legal target the controller is asked, as it resolves, whether Kaya or the
// target goes. The exiled card comes back at the beginning of the controller's
// next upkeep as a new object (CR 400.7), so Kaya returns with her printed
// loyalty. You lose 2 life either way.
//
// No simplification.
func init() {
	kayaGhostAssassinPick = effects.OptionPickThen("option-pick/kaya-ghost-assassin-exile", kayaGhostAssassinPicked)
	effects.Register(effects.Spec{
		OracleID:        "5e4c81b9-a31e-45e6-983f-30f11b8a05c0",
		Name:            "Kaya, Ghost Assassin",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "0: Exile Kaya or up to one target creature. Return that card to the battlefield under its owner's control at the beginning of your next upkeep. You lose 2 life.",
				Cost:    effects.LoyaltyCost(0),
				Targets: effects.TargetCreature("up to one target creature").WithCount(0, 1),
				Effect:  kayaGhostAssassinZero,
			},
			{
				Label: "−1: Each opponent loses 2 life and you gain 2 life.",
				Cost:  effects.LoyaltyCost(-1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.EachOpponentLosesTwoYouGainTwo(item, effects.NewContext(g, item), 0)
				},
			},
			{
				Label: "−2: Each opponent discards a card and you draw a card.",
				Cost:  effects.LoyaltyCost(-2),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if err := effects.EachOpponentDiscardsOne(g, item); err != nil {
						return err
					}
					return effects.DrawCards{Player: item.Controller, N: 1}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}

// kayaGhostAssassinPick is the registered continuation of the 0's question. An
// on-disk key: never renamed, never reused.
var kayaGhostAssassinPick game.OptionPickThen

func kayaGhostAssassinZero(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), item.Controller, -2); err != nil {
		return err
	}
	var target uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			target = t.ID
			break
		}
	}
	if target == uuid.Nil {
		return kayaGhostAssassinExile(ctx, item.SourceCardID)
	}
	name := "the target creature"
	if c, ok := g.LookupCardForEffect(target); ok {
		name = c.Name
	}
	return effects.PickOption{
		Question: "Kaya, Ghost Assassin — exile Kaya or " + name + "?",
		Options:  []game.ChoiceOption{{Label: "Exile " + name}, {Label: "Exile Kaya"}},
		ThenKey:  kayaGhostAssassinPick,
		Carry:    []uuid.UUID{target, item.SourceCardID},
	}.Apply(ctx)
}

func kayaGhostAssassinPicked(ctx *effects.Context, r game.OptionPicked) error {
	if r.Option == nil || len(r.Carry) < 2 {
		return nil
	}
	if r.Index == 0 {
		return kayaGhostAssassinExile(ctx, r.Carry[0])
	}
	return kayaGhostAssassinExile(ctx, r.Carry[1])
}

// kayaGhostAssassinExile exiles the card (if it is still on the battlefield)
// and schedules its return at the beginning of its controller's next upkeep.
func kayaGhostAssassinExile(ctx *effects.Context, id uuid.UUID) error {
	if id == uuid.Nil || !effects.OnBattlefield(ctx.Game, id) {
		return nil
	}
	return ctx.Game.ExileCardsKeyedForEffect([]uuid.UUID{id}, ctx.ItemTail(kayaGhostAssassinSchedule))
}

var kayaGhostAssassinSchedule = effects.CardTail("kaya-ghost-assassin/schedule-upkeep-return", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return effects.ScheduleDelayedTrigger{
		At:                 game.StepUpkeep,
		ControllerTurnOnly: true,
		Label:              "Kaya, Ghost Assassin — return the exiled card to the battlefield under its owner's control",
		Cards:              o.IDs,
		Body:               effects.ReturnExiledToOwnersBody,
	}.Apply(ctx)
})
