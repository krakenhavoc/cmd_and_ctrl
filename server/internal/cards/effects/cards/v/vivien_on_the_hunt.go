package cardsv

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vivien on the Hunt — Legendary Planeswalker — Vivien {4}{G}{G}, loyalty 4
// (Streets of New Capenna, #5491):
//
//	"+2: You may sacrifice a creature. If you do, search your library for a
//	 creature card with mana value equal to 1 plus the sacrificed creature's
//	 mana value, put it onto the battlefield, then shuffle.
//	 +1: Mill five cards, then put any number of creature cards milled this
//	 way into your hand.
//	 −1: Create a 4/4 green Rhino Warrior creature token."
//
// The +2 asks as it resolves (CR 608.2c) and the search runs only for a
// sacrifice that happened ("if you do"), against the sacrificed creature's
// last-known mana value (CR 608.2h; a token's is zero, a copy's is the
// copied card's). The +1 picks over the cards that really reached the
// graveyard (CR 400.7), any number of them, none included.
//
// No simplification.
func init() {
	const name = "Vivien on the Hunt"
	effects.Register(effects.Spec{
		OracleID:        "f6f927f3-dc3f-446c-aa7b-81048eb73243",
		Name:            name,
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 4,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+2: You may sacrifice a creature. If you do, search your library for a creature card with mana value equal to 1 plus the sacrificed creature's mana value, put it onto the battlefield, then shuffle.",
				Purpose: game.Purpose{Tutors: 1},
				Cost:    effects.LoyaltyCost(2),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.MaySacrificeOne(g, item, name+" — sacrifice a creature to search for a creature with one more mana value?",
						effects.Creature(), false, vivienOnTheHuntSacrificed)
				},
			},
			{
				Label: "+1: Mill five cards, then put any number of creature cards milled this way into your hand.",
				Cost:  effects.LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return g.MillToZoneKeyedForEffect(ctx.Controller(), 5, game.ZoneGraveyard, game.MillUntil{},
						effects.WithWord(ctx.ItemTail(vivienOnTheHuntMilled), name+" — put any number of creature cards milled this way into your hand"))
				},
			},
			{
				Label:   "−1: Create a 4/4 green Rhino Warrior creature token.",
				Purpose: game.Purpose{Tokens: 1},
				Cost:    effects.LoyaltyCost(-1),
				Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("4/4 green Rhino Warrior"), N: 1}),
			},
		},
	})
}

// vivienOnTheHuntSacrificed is the "if you do": the search for one more mana
// value, onto the battlefield. An on-disk key: never renamed, never reused.
var vivienOnTheHuntSacrificed = effects.RegisterSacrificedOne("prompt-run/vivien-on-the-hunt-sacrificed", func(ctx *effects.Context, sacrificed uuid.UUID) error {
	want := effects.DepartedManaValue(ctx.Game, sacrificed) + 1
	return effects.SearchLibrary{
		Player:    ctx.Controller(),
		Predicate: func(c game.Card) bool { return c.IsCreature() && c.ManaValue() == want },
		Dest:      game.ZoneBattlefield,
		Limit:     1,
		Shuffle:   true,
		Reason:    "Vivien on the Hunt — a creature card with the sacrificed creature's mana value plus one",
	}.Apply(ctx)
})

// vivienOnTheHuntPick puts the chosen milled creature cards into your hand.
var vivienOnTheHuntPick = effects.AskedPick("choose-cards/"+strings.ReplaceAll("mill-then-take/vivien-on-the-hunt", "/", "-")+"-pick", func(ctx *effects.Context, r game.CardsChosen) error {
	for _, id := range r.Picked {
		if err := (effects.ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})

// vivienOnTheHuntMilled is the mill's keyed "then": the pick over the
// creature cards that reached the graveyard. An on-disk key: never renamed,
// never reused.
var vivienOnTheHuntMilled = effects.CardTail("mill-then-take/vivien-on-the-hunt", func(ctx *effects.Context, o game.TailOutcome) error {
	var eligible []uuid.UUID
	for _, id := range o.IDs {
		if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && c.IsCreature() && !effects.IsToken(c) {
			eligible = append(eligible, id)
		}
	}
	if len(eligible) == 0 {
		return nil
	}
	item := ctx.Item
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: o.Tail.Word,
		Cards:    eligible,
		Min:      0,
		Max:      len(eligible),
		LiveZone: game.ZoneGraveyard,
		ThenKey:  vivienOnTheHuntPick,
		Asked:    game.NewAskedItem(item),
	})
	return nil
})
