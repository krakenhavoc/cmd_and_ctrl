package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Charnel Troll — Creature — Troll {1}{B}{G}, 4/4 (Guilds of Ravnica):
//
//	"Trample
//	 At the beginning of your upkeep, exile a creature card from your
//	 graveyard. If you do, put a +1/+1 counter on this creature. Otherwise,
//	 sacrifice it.
//	 {B}{G}, Discard a creature card: Put a +1/+1 counter on this creature."
//
// The upkeep exile is not optional: with a creature card in the graveyard its
// controller must choose one (a one-card prompt, never a "may"), and with none
// the Troll is sacrificed. The counter goes on only if the card really reached
// exile (a replacement may send it elsewhere, and then it is "otherwise"), and
// only while the Troll is still the same permanent. The activated row's
// discard is a cost paid at activation (Fauna Shaman).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c7c35dd9-8c3b-4be5-b310-61db5e73ed6c",
		Name:            "Charnel Troll",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Charnel Troll — exile a creature card from your graveyard; if you do, put a +1/+1 counter on it, otherwise sacrifice it",
				charnelTrollUpkeep),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}{G}, Discard a creature card: Put a +1/+1 counter on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump, SourceCounters: 1},
			Cost:    effects.Plus(effects.ManaCost("{B}{G}"), effects.DiscardCardsMatching(1, "a creature card", effects.MatchCreature)),
			Effect:  effects.PutACounterOnThis(game.CounterPlusOne),
		}},
	})
}

// charnelTrollUpkeep asks which creature card to exile, or sacrifices the
// Troll when its controller's graveyard has none.
func charnelTrollUpkeep(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if !effects.OnBattlefield(g, item.SourceCardID) || effects.SourceIsNewObject(g, item) {
		return nil
	}
	cards := effects.GraveyardCardIDs(ctx, ctx.Controller(), func(c game.Card) bool { return c.IsCreature() })
	if len(cards) == 0 {
		return effects.SacrificePermanent{Target: item.SourceCardID}.Apply(ctx)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    ctx.Controller(),
		FromPlayer: ctx.Controller(),
		Source:     item.SourceCardID,
		Question:   "Charnel Troll — exile a creature card from your graveyard",
		Cards:      cards,
		Min:        1,
		Max:        1,
		LiveZone:   game.ZoneGraveyard,
		Asked:      game.NewAskedItem(item),
		ThenKey:    charnelTrollPicked,
	})
	return nil
}

// charnelTrollPicked exiles the chosen card; the clause after it decides
// between the counter and the sacrifice. An on-disk key.
var charnelTrollPicked = effects.AskedPick("choose-cards/charnel-troll-exile", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) != 1 {
		return charnelTrollOtherwise(ctx)
	}
	return effects.ExileTarget{Target: r.Picked[0], ThenTail: ctx.ItemTail(charnelTrollExiled)}.Apply(ctx)
})

// charnelTrollExiled is "if you do, put a +1/+1 counter on this creature.
// Otherwise, sacrifice it." An on-disk key.
var charnelTrollExiled = effects.CardTail("charnel-troll/after-exile", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) != 1 {
		return charnelTrollOtherwise(ctx)
	}
	if !effects.OnBattlefield(ctx.Game, ctx.Source()) || effects.SourceIsNewObject(ctx.Game, ctx.Item) {
		return nil
	}
	return effects.AddCounter{Target: ctx.Source(), Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
})

func charnelTrollOtherwise(ctx *effects.Context) error {
	if !effects.OnBattlefield(ctx.Game, ctx.Source()) || effects.SourceIsNewObject(ctx.Game, ctx.Item) {
		return nil
	}
	return effects.SacrificePermanent{Target: ctx.Source()}.Apply(ctx)
}
