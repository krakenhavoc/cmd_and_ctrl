package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Resplendent Marshal — Creature — Angel Warrior {1}{W}{W}, 3/3 (Kaldheim, #5505):
//
//	"Flying
//	 When this creature enters or dies, you may exile another creature
//	 card from your graveyard. When you do, put a +1/+1 counter on each
//	 creature you control other than this creature that shares a creature
//	 type with the exiled card."
//
// One ability, two conditions (WhenThisEntersOrDies). The "you may" is a
// pick of up to one creature card, never the Marshal itself (it sits in
// the graveyard when the dies half resolves). The exile is the "if you
// do": the reflexive trigger (CR 603.12) goes on the stack above only if a
// card was exiled, and reads its creature types from that card in exile.
// A changeling shares a type with every creature that has one
// (game.SharesCreatureType). Creatures are counted when it resolves.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1fb4c1c1-b10c-463f-a97d-eca9f978e371",
		Name:            "Resplendent Marshal",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersOrDies("Resplendent Marshal — you may exile another creature card from your graveyard; when you do, put a +1/+1 counter on each other creature you control that shares a creature type with it",
				resplendentMarshalAsk),
		},
	})
}

func resplendentMarshalAsk(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	candidates := effects.ResplendentMarshalCandidates(ctx, item.Controller, item.SourceCardID)
	if len(candidates) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    item.Controller,
		FromPlayer: item.Controller,
		Source:     item.SourceCardID,
		Question:   "Resplendent Marshal — you may exile another creature card from your graveyard",
		Cards:      candidates,
		Min:        0,
		Max:        1,
		LiveZone:   game.ZoneGraveyard,
		ThenKey:    resplendentMarshalPick,
		Asked:      game.NewAskedItem(item),
	})
	return nil
}

var resplendentMarshalPick = effects.AskedPick("choose-cards/resplendent-marshal-exile", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return effects.ExileTarget{Target: r.Picked[0], ThenTail: ctx.ItemTail(resplendentMarshalExiled)}.Apply(ctx)
})

var resplendentMarshalExiled = effects.CardTail("resplendent-marshal/exiled", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return effects.ReflexiveTrigger{
		Label: "Resplendent Marshal — put a +1/+1 counter on each other creature you control that shares a creature type with the exiled card",
		Cards: []uuid.UUID{o.IDs[0]},
		Body:  resplendentMarshalCountersBody,
	}.Apply(ctx)
})

var resplendentMarshalCountersBody = game.ReflexiveBody("resplendent-marshal/counters", effects.SimpleBody(resplendentMarshalCounters), nil)

func resplendentMarshalCounters(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	exiled := ctx.PayloadCards()
	if len(exiled) == 0 {
		return nil
	}
	card, ok := g.LookupCardForEffect(exiled[0])
	if !ok {
		return nil
	}
	var ids []uuid.UUID
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == item.Controller && c.InstanceID != item.SourceCardID && c.IsCreature() && game.SharesCreatureType(c, &card) {
			ids = append(ids, c.InstanceID)
		}
	}
	for _, id := range ids {
		if err := (effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx.AsGroupMember()); err != nil {
			return err
		}
	}
	return nil
}
