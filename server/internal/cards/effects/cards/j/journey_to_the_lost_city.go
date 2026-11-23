package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Journey to the Lost City — Enchantment {3}{G} (Commander Legends: Battle
// for Baldur's Gate, #5948):
//
//	"At the beginning of your upkeep, exile the top four cards of your
//	 library, then roll a d20.
//	 1—9 | You may put a land card from among those cards onto the
//	 battlefield.
//	 10—19 | Create a 2/2 green Wolf creature token, then put a +1/+1 counter
//	 on it for each creature card among those cards.
//	 20 | Put all permanent cards exiled with this enchantment onto the
//	 battlefield, then sacrifice it."
//
// The four cards are exiled as one plain exile (ExileTopKeyedForEffect), and
// the d20 is rolled once they have all settled, the table reading them as
// data. 1—9 asks which one land card from among them (if any) to put onto the
// battlefield; the others stay exiled, and so does a land you decline. 10—19
// makes the Wolf and then, as separate actions (CR 608.2c), puts a counter on
// it for each creature card among the four (a Wolf that has already gone gets
// none). 20 reads every card this enchantment has exiled and that is still in
// exile, from this and every earlier upkeep (B27ExiledWith, the Oblivion Ring
// record, which is reset if the enchantment leaves and comes back, CR 400.7),
// puts the permanent cards among them onto the battlefield together under
// their owners' control, then sacrifices the enchantment.
//
// No simplification.
const journeyLabel = "Journey to the Lost City — exile the top four cards of your library, then roll a d20"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "a5f5e662-b32c-44f1-848d-0228f0fe1eb6",
		Name:         "Journey to the Lost City",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep(journeyLabel, func(g *game.Game, item *game.StackItem) error {
				return g.ExileTopKeyedForEffect(item.Controller, item.Controller, 4, effects.NewContext(g, item).ItemTail(journeyExiled))
			}),
		},
	})
}

// journeyExiled rolls the d20 once the four cards are in exile. An on-disk
// key: never renamed, never reused.
var journeyExiled = effects.CardTail("journey-to-the-lost-city/exiled", func(ctx *effects.Context, o game.TailOutcome) error {
	return effects.RollDice{Sides: 20, N: 1, Then: journeyTable, TailCarry: o.IDs}.Apply(ctx)
})

// journeyTable is the results table, told the four cards as Tail.Carry. An
// on-disk key.
var journeyTable = effects.CardTail("journey-to-the-lost-city/table", func(ctx *effects.Context, o game.TailOutcome) error {
	g, you := ctx.Game, ctx.Controller()
	roll := effects.DiceResult(o)
	inExile := func(id uuid.UUID, match func(game.Card) bool) bool {
		c, ok := g.LookupCardForEffect(id)
		if !ok || c.IsToken() || !match(c) {
			return false
		}
		z := g.FindCardZoneForEffect(id)
		return z != nil && z.Kind == game.ZoneExile
	}
	switch {
	case roll >= 1 && roll <= 9:
		var lands []uuid.UUID
		for _, id := range o.Tail.Carry {
			if inExile(id, func(c game.Card) bool { return c.IsLand() }) {
				lands = append(lands, id)
			}
		}
		if len(lands) == 0 {
			return nil
		}
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:  you,
			Source:   ctx.Source(),
			Question: "Journey to the Lost City — you may put a land card from among those cards onto the battlefield",
			Cards:    lands,
			Min:      0,
			Max:      1,
			LiveZone: game.ZoneExile,
			ThenKey:  journeyLandPick,
			Asked:    game.NewAskedItem(ctx.Item),
		})
		return nil
	case roll >= 10 && roll <= 19:
		creatures := 0
		for _, id := range o.Tail.Carry {
			if inExile(id, func(c game.Card) bool { return c.IsCreature() }) {
				creatures++
			}
		}
		cursor := effects.B25LastEventSeq(g)
		if err := (effects.CreateToken{Template: effects.TokenCard("2/2 green Wolf"), N: 1}).Apply(ctx); err != nil {
			return err
		}
		if creatures == 0 {
			return nil
		}
		for _, token := range effects.B27TokensCreatedByAfter(g, you, cursor) {
			return effects.AddCounter{Target: token, Kind: game.CounterPlusOne, N: creatures}.Apply(ctx)
		}
		return nil
	case roll == 20:
		var permanents []uuid.UUID
		for _, id := range effects.B27ExiledWith(g, ctx.Source(), journeyLabel) {
			if c, ok := g.LookupCardForEffect(id); ok && !c.IsToken() && !c.IsInstant() && !c.IsSorcery() {
				permanents = append(permanents, id)
			}
		}
		if len(permanents) > 0 {
			if err := (effects.ReturnFromExileTogether{Targets: permanents}).Apply(ctx); err != nil {
				return err
			}
		}
		return effects.SacrificePermanent{Target: ctx.Source()}.Apply(ctx)
	}
	return nil
})

// journeyLandPick puts the chosen land onto the battlefield. An on-disk key.
var journeyLandPick = effects.AskedPick("choose-cards/journey-to-the-lost-city-land", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return effects.ReturnFromExileTogether{Targets: r.Picked, Controller: ctx.Controller()}.Apply(ctx)
})
