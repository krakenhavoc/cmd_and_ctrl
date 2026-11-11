package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Winter, Cynical Opportunist — Legendary Creature — Human Warlock {2}{B}{G},
// 2/5 (Duskmourn Commander):
//
//	"Deathtouch
//	 Whenever Winter attacks, mill three cards.
//	 Delirium — At the beginning of your end step, you may exile any number of
//	 cards from your graveyard with four or more card types among them. If you
//	 do, put a permanent card from among them onto the battlefield with a
//	 finality counter on it."
//
// The end-step ability is two keyed choose-cards prompts. The first asks, as
// the trigger resolves (CR 603.5), which graveyard cards to exile; its answer
// is checked against a registered set rule, "four or more card types among
// the picked cards" (cmdr2S29FourCardTypes), so a smaller set is not an answer
// the engine accepts or the enumerator offers, and choosing none is the "may"
// declined. No prompt is asked when the whole graveyard holds fewer than four
// card types, since no set could qualify. "If you do" is the exile having
// happened: the picked cards that reached exile are the pool of the second
// prompt, a mandatory pick of one permanent card from among them, which enters
// under your control with a finality counter (CR 122.1h) as one entry.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e70a01dd-4dbe-4811-ac93-1c6d0142e06e",
		Name:            "Winter, Cynical Opportunist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Winter, Cynical Opportunist — mill three cards", Do(MillCards{N: 3})),
			AtYourEndStep("Winter, Cynical Opportunist — you may exile any number of cards from your graveyard with four or more card types among them",
				winterEndStep),
		},
	})
}

func winterEndStep(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	cands := graveyardCardIDs(ctx, item.Controller, func(game.Card) bool { return true })
	if len(cands) == 0 || b16CardTypesInGraveyard(g, item.Controller) < 4 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    item.Controller,
		FromPlayer: item.Controller,
		Source:     item.SourceCardID,
		Question:   "Winter, Cynical Opportunist — you may exile any number of cards from your graveyard with four or more card types among them",
		Cards:      cands,
		Min:        0,
		Max:        0,
		ThenKey:    winterExile,
		LiveZone:   game.ZoneGraveyard,
	})
	return nil
}

// winterExile is the keyed answer to "which cards to exile". An on-disk key:
// never renamed, never reused.
var winterExile = game.RegisterChooseCardsThenRule("choose-cards/winter-cynical-opportunist-exile", cmdr2S29FourCardTypes,
	func(g *game.Game, r game.CardsChosen) error {
		if !r.Answered || len(r.Picked) == 0 {
			return nil
		}
		var ids []uuid.UUID
		for _, id := range r.Picked {
			if z := g.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneGraveyard {
				ids = append(ids, id)
			}
		}
		if len(ids) == 0 {
			return nil
		}
		return g.ExileCardsKeyedForEffect(ids, game.KeyedTail{Ref: winterExiled, Player: r.Chooser, Source: r.Source})
	})

// winterExiled is "if you do" and "from among them": told the cards that
// reached exile (as they now are, CR 400.7), it asks for one permanent card
// among them. An on-disk key.
var winterExiled = CardTail("tail/winter-cynical-opportunist-exiled", func(ctx *Context, o game.TailOutcome) error {
	var permanents []uuid.UUID
	for _, id := range o.IDs {
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && c.IsPermanent() {
			permanents = append(permanents, id)
		}
	}
	if len(permanents) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  o.Tail.Player,
		Source:   o.Tail.Source,
		Question: "Winter, Cynical Opportunist — put a permanent card from among the exiled cards onto the battlefield with a finality counter on it",
		Cards:    permanents,
		Min:      1,
		Max:      1,
		ThenKey:  winterPut,
		LiveZone: game.ZoneExile,
	})
	return nil
})

// winterPut is the keyed answer to "which permanent card": it enters under
// the chooser's control with a finality counter. An on-disk key.
var winterPut = game.RegisterChooseCardsThen("choose-cards/winter-cynical-opportunist-put", func(g *game.Game, r game.CardsChosen) error {
	if !r.Answered || len(r.Picked) == 0 {
		return nil
	}
	id := r.Picked[0]
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile {
		return nil
	}
	return g.ReturnFromExileWithCountersKeyedForEffect(id, r.Chooser, false,
		map[string]int{game.CounterFinality: 1}, game.KeyedTail{})
})
