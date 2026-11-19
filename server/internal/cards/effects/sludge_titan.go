package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sludge Titan — Creature — Zombie Giant {4}{B/G}{B/G}, 6/6 (Ravnica: Clue
// Edition):
//
//	"Trample
//	 Whenever this creature enters or attacks, mill five cards. You may put
//	 a creature card and/or a land card from among them into your hand."
//
// Lluwen's mill with a keyed pick over the milled cards still in the
// graveyard (CR 400.7), asked twice: first an optional creature card,
// then an optional land card from the rest. Asking in sequence is the
// same choice as "a creature card and/or a land card".
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "94dd397e-0127-4c79-b9d9-22ce62ccdfe7",
		Name:            "Sludge Titan",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks("Sludge Titan — mill five cards, you may put a creature card and/or a land card from among them into your hand",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return g.MillToZoneKeyedForEffect(item.Controller, 5, game.ZoneGraveyard, game.MillUntil{}, ctx.ItemTail(sludgeTitanMilled))
				}),
		},
	})
}

// sludgeTitanMatching is the milled cards still in the graveyard that
// match.
func sludgeTitanMatching(ctx *Context, ids []uuid.UUID, match func(c game.Card) bool) []uuid.UUID {
	var out []uuid.UUID
	for _, id := range ids {
		if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && match(c) {
			out = append(out, id)
		}
	}
	return out
}

// sludgeTitanAsk queues one optional pick; the milled ids ride as Carry.
func sludgeTitanAsk(ctx *Context, cands, carry []uuid.UUID, question string, then game.ChooseCardsThen) {
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: question,
		Cards:    cands,
		Min:      0,
		Max:      1,
		LiveZone: game.ZoneGraveyard,
		Carry:    carry,
		ThenKey:  then,
		Asked:    game.NewAskedItem(ctx.Item),
	})
}

const (
	sludgeTitanCreatureQ = "Sludge Titan — you may put a creature card milled this way into your hand"
	sludgeTitanLandQ     = "Sludge Titan — you may put a land card milled this way into your hand"
)

// sludgeTitanMilled is the mill's keyed "then": offer a creature card
// first, or go straight to the land pick when none was milled. An
// on-disk key: never renamed, never reused.
var sludgeTitanMilled = CardTail("mill-then/sludge-titan", func(ctx *Context, o game.TailOutcome) error {
	creatures := sludgeTitanMatching(ctx, o.IDs, func(c game.Card) bool { return c.IsCreature() })
	if len(creatures) > 0 {
		sludgeTitanAsk(ctx, creatures, o.IDs, sludgeTitanCreatureQ, sludgeTitanCreaturePick)
		return nil
	}
	if lands := sludgeTitanMatching(ctx, o.IDs, func(c game.Card) bool { return c.IsLand() }); len(lands) > 0 {
		sludgeTitanAsk(ctx, lands, o.IDs, sludgeTitanLandQ, sludgeTitanLandPick)
	}
	return nil
})

var sludgeTitanCreaturePick = askedPick("choose-cards/sludge-titan-creature", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) > 0 {
		if err := (ReturnFromGraveyard{Target: r.Picked[0], Dest: game.ZoneHand}).Apply(ctx); err != nil {
			return err
		}
	}
	if lands := sludgeTitanMatching(ctx, r.Carry, func(c game.Card) bool { return c.IsLand() }); len(lands) > 0 {
		sludgeTitanAsk(ctx, lands, r.Carry, sludgeTitanLandQ, sludgeTitanLandPick)
	}
	return nil
})

var sludgeTitanLandPick = askedPick("choose-cards/sludge-titan-land", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return ReturnFromGraveyard{Target: r.Picked[0], Dest: game.ZoneHand}.Apply(ctx)
})
