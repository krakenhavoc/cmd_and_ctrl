package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Skywarp Skaab — Creature — Zombie Drake {3}{U}{U}, 2/5 (Innistrad: Crimson
// Vow):
//
//	"Flying
//	 When this creature enters, you may exile two creature cards from your
//	 graveyard. If you do, draw a card."
//
// "You may exile two" is a yes/no first, then exactly two creature cards
// picked as the trigger resolves (a graveyard with fewer than two does
// nothing, as the exile cannot be performed). The draw is told the cards that
// really reached exile and happens only when both did (CR 608.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e11b2b6e-3819-4dcb-b64e-2dfe6a6b049d",
		Name:            "Skywarp Skaab",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			Optional(WhenThisEnters("Skywarp Skaab — you may exile two creature cards from your graveyard; if you do, draw a card",
				skywarpSkaabOffer), "Skywarp Skaab — exile two creature cards from your graveyard to draw a card?"),
		},
	})
}

func skywarpSkaabOffer(g *game.Game, item *game.StackItem) error {
	ids := b42CardsInGraveyardOf(g, item.Controller, func(c game.Card) bool { return c.IsCreature() })
	if len(ids) < 2 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Skywarp Skaab — choose two creature cards from your graveyard to exile (you draw a card)",
		Cards:    ids,
		Min:      2,
		Max:      2,
		LiveZone: game.ZoneGraveyard,
		ThenKey:  skywarpSkaabExile,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// skywarpSkaabExile exiles the two cards picked. An on-disk key.
var skywarpSkaabExile = askedPick("choose-cards/skywarp-skaab-exile", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) != 2 {
		return nil
	}
	return ctx.Game.ExileCardsKeyedForEffect(r.Picked, ctx.ItemTail(skywarpSkaabDraw))
})

// skywarpSkaabDraw is "if you do, draw a card", told the cards that reached
// exile. An on-disk key.
var skywarpSkaabDraw = CardTail("exile-then/skywarp-skaab-draw", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) != 2 {
		return nil
	}
	return DrawCards{N: 1}.Apply(ctx)
})
