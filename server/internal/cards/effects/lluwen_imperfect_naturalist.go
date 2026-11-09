package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lluwen, Imperfect Naturalist — Legendary Creature — Elf Druid {B/G}{B/G}
// (Lorwyn Eclipsed, #4458):
//
//	"When Lluwen enters, mill four cards, then you may put a creature or
//	 land card from among the milled cards on top of your library.
//	 {2}{B/G}{B/G}{B/G}, {T}, Discard a land card: Create a 1/1 black and
//	 green Worm creature token for each land card in your graveyard."
//
// The pick is over the milled cards still in the graveyard (CR 400.7). The
// discarded land is a cost, so it is in the graveyard when the tokens are
// counted and counts itself.
func init() {
	Register(Spec{
		OracleID:     "81ff5764-72ce-4b9a-a14f-07413700d935",
		Name:         "Lluwen, Imperfect Naturalist",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Lluwen, Imperfect Naturalist — mill four cards, then you may put a creature or land card from among them on top of your library",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return g.MillToZoneKeyedForEffect(item.Controller, 4, game.ZoneGraveyard, game.MillUntil{}, ctx.ItemTail(lluwenMilled))
				}),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}{B/G}{B/G}{B/G}, {T}, Discard a land card: Create a 1/1 black and green Worm creature token for each land card in your graveyard.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost: Plus(ManaCost("{2}{B/G}{B/G}{B/G}"), TapCost(),
				DiscardCardsMatching(1, "a land card", func(c game.Card) bool { return c.IsLand() })),
			Effect: func(g *game.Game, item *game.StackItem) error {
				n := len(b42CardsInGraveyardOf(g, item.Controller, func(c game.Card) bool { return c.IsLand() }))
				return CreateToken{Template: TokenCard("1/1 black and green Worm"), N: n}.Apply(NewContext(g, item))
			},
		}},
	})
}

// lluwenMilled is the mill's keyed "then": offer a creature or land card
// among the milled ones. An on-disk key: never renamed, never reused.
var lluwenMilled = CardTail("mill-then/lluwen-imperfect-naturalist", func(ctx *Context, o game.TailOutcome) error {
	var eligible []uuid.UUID
	for _, id := range o.IDs {
		if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && (c.IsCreature() || c.IsLand()) {
			eligible = append(eligible, id)
		}
	}
	if len(eligible) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Lluwen, Imperfect Naturalist — you may put a creature or land card milled this way on top of your library",
		Cards:    eligible,
		Min:      0,
		Max:      1,
		LiveZone: game.ZoneGraveyard,
		ThenKey:  lluwenPick,
		Asked:    game.NewAskedItem(ctx.Item),
	})
	return nil
})

var lluwenPick = askedPick("choose-cards/lluwen-imperfect-naturalist-top", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return ReturnFromGraveyard{Target: r.Picked[0], Dest: game.ZoneLibrary}.Apply(ctx)
})
