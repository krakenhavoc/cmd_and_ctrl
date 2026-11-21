package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Eivor, Wolf-Kissed — Legendary Creature — Human Assassin Warrior
// {3}{R}{G}{W}, 7/6 (Assassin's Creed, #5948):
//
//	"Trample, haste
//	 Whenever Eivor deals combat damage to a player, you mill that many
//	 cards. You may put a Saga card and/or a land card from among them onto
//	 the battlefield."
//
// The amount is the combat damage the event carries. Bristlebud Farmer's
// mill-then-take: only cards this mill put into the graveyard (and still
// there) are offered, in one optional pick of up to two cards, at most one
// Saga card and at most one land card (a set rule, so a pick of two Sagas is
// refused), each put onto the battlefield untapped under your control.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "cb8289e9-1e59-41f1-a994-6e34d2da2da2",
		Name:            "Eivor, Wolf-Kissed",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, ThisDealtCombatDamageToAPlayer,
				"Eivor, Wolf-Kissed — mill that many cards; you may put a Saga card and/or a land card from among them onto the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return MillToZone{N: ctx.Trigger().Event.Amount, ThenTail: ctx.ItemTail(eivorMilled)}.Apply(ctx)
				}),
		},
	})
}

// eivorSagaOrLand is a card of the kinds Eivor puts onto the battlefield.
func eivorSagaOrLand(c game.Card) (saga, land bool) {
	return c.HasSubtype("Saga"), c.IsLand()
}

// eivorMilled is the tail told every card the mill put into a graveyard.
// An on-disk key.
var eivorMilled = CardTail("eivor-wolf-kissed/after-mill", func(ctx *Context, o game.TailOutcome) error {
	var candidates []uuid.UUID
	for _, id := range o.IDs {
		z := ctx.Game.FindCardZoneForEffect(id)
		if z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		if c, ok := ctx.Game.LookupCardForEffect(id); ok {
			if saga, land := eivorSagaOrLand(c); saga || land {
				candidates = append(candidates, id)
			}
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Eivor, Wolf-Kissed — you may put a Saga card and/or a land card from among the milled cards onto the battlefield",
		Cards:    candidates,
		Min:      0,
		Max:      2,
		ThenKey:  eivorPicked,
		Asked:    game.NewAskedItem(ctx.Item),
	})
	return nil
})

// eivorPicked puts the chosen cards onto the battlefield. At most one Saga
// and at most one land. An on-disk key.
var eivorPicked = askedPickRule("choose-cards/eivor-wolf-kissed-put",
	func(picked []game.Card) bool {
		sagas, lands := 0, 0
		for _, c := range picked {
			saga, land := eivorSagaOrLand(c)
			if saga {
				sagas++
			}
			if land {
				lands++
			}
			if !saga && !land {
				return false
			}
		}
		return sagas <= 1 && lands <= 1
	},
	func(ctx *Context, r game.CardsChosen) error {
		for _, id := range r.Picked {
			if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
				continue
			}
			if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield, Controller: ctx.Controller()}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
