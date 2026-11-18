package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stitcher Geralf — Legendary Creature — Human Wizard {3}{U}{U}, 3/4:
//
//	"{2}{U}, {T}: Each player mills three cards. Exile up to two creature
//	 cards put into graveyards this way. Create an X/X blue Zombie
//	 creature token, where X is the total power of the cards exiled this
//	 way."
//
// Each library is milled in APNAP order (millEachPlayerInTurn); the
// cards that actually reached a graveyard are carried forward by value
// (Dread Summons' chain). The controller then picks up to two of the
// creature cards among them that are still in a graveyard, those are
// exiled together, and X is their total power as they stand in exile
// (a characteristic-defining power counts, a card with no power counts
// as 0). The token is made even when X is 0 (it dies as a 0/0), as the
// text says.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "da394a25-79c9-4b75-9ce6-741f64ae6913",
		Name:         "Stitcher Geralf",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Label:   "{2}{U}, {T}: Each player mills three cards. Exile up to two creature cards put into graveyards this way. Create an X/X blue Zombie creature token, where X is the total power of the cards exiled this way.",
			Cost:    Plus(ManaCost("{2}{U}"), TapCost()),
			Effect:  stitcherGeralfMill,
		}},
	})
}

func stitcherGeralfMill(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var seats []uuid.UUID
	for _, p := range g.Seats {
		if p != nil {
			seats = append(seats, p.ID)
		}
	}
	return millEachPlayerInTurn(ctx, seats, constantMill(3), ctx.ItemTail(stitcherGeralfAsk))
}

// stitcherGeralfAsk is told every card the mills put into a graveyard.
// An on-disk key: never renamed, never reused.
var stitcherGeralfAsk = CardTail("stitcher-geralf/ask", func(ctx *Context, o game.TailOutcome) error {
	var candidates []uuid.UUID
	for _, id := range o.IDs {
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && c.IsCreature() && cardInAGraveyard(ctx.Game, id) {
			candidates = append(candidates, id)
		}
	}
	if len(candidates) == 0 {
		return stitcherGeralfZombie(ctx, 0)
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Stitcher Geralf — exile up to two creature cards put into graveyards this way",
		Cards:    candidates,
		Min:      0,
		Max:      min(2, len(candidates)),
		ThenKey:  stitcherGeralfPick,
		Asked:    game.NewAskedItem(ctx.Item),
	})
	return nil
})

// stitcherGeralfPick exiles the picked cards together. An on-disk key.
var stitcherGeralfPick = askedPick("choose-cards/stitcher-geralf-exile", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return stitcherGeralfZombie(ctx, 0)
	}
	return ctx.Game.ExileCardsKeyedForEffect(r.Picked, ctx.ItemTail(stitcherGeralfExiled))
})

// stitcherGeralfExiled is told the cards that reached exile and makes the
// Zombie. An on-disk key.
var stitcherGeralfExiled = CardTail("stitcher-geralf/exiled", func(ctx *Context, o game.TailOutcome) error {
	x := 0
	for _, id := range o.IDs {
		if c, ok := ctx.Game.LookupCardForEffect(id); ok {
			x += c.PowerForComparison()
		}
	}
	return stitcherGeralfZombie(ctx, x)
})

func stitcherGeralfZombie(ctx *Context, x int) error {
	tok := ctx.TextToken(TokenCard("0/0 blue Zombie"))
	tok.Power, tok.Toughness = max(x, 0), max(x, 0)
	return ctx.Game.CreateTokenForEffect(ctx.Controller(), tok, 1)
}

// cardInAGraveyard reports whether the card is in some player's graveyard.
func cardInAGraveyard(g *game.Game, id uuid.UUID) bool {
	z := g.FindCardZoneForEffect(id)
	return z != nil && z.Kind == game.ZoneGraveyard
}
