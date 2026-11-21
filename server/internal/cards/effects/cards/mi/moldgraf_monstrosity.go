package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Moldgraf Monstrosity — Creature — Insect {4}{G}{G}{G}, 8/8 (Duskmourn
// Commander):
//
//	"Trample
//	 When this creature dies, exile it, then return two creature cards at
//	 random from your graveyard to the battlefield."
//
// The trigger exiles the card if it is still in the graveyard (Ancient
// Adamantoise's shape), then, whether or not that worked, picks up to two
// creature cards at random from your graveyard as the rest of the clause runs
// (ADR 0054's keyed stream, so an undo or restore replays the same pick). The
// Monstrosity is out of the graveyard by then and cannot return itself; with
// a single creature card the one is returned, with none nothing is. The two
// enter as one simultaneous entry.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3c4d5392-27cf-492e-9ae1-dee330b82223",
		Name:            "Moldgraf Monstrosity",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Moldgraf Monstrosity — exile it, then return two creature cards at random from your graveyard to the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneGraveyard {
						return moldgrafReturnTwoAtRandom(ctx)
					}
					return effects.ExileTarget{
						Target:   item.SourceCardID,
						ThenTail: ctx.ItemTail(moldgrafMonstrosityExiled),
					}.Apply(ctx)
				}),
		},
	})
}

// moldgrafMonstrosityExiled is the clause after the exile. An on-disk key.
var moldgrafMonstrosityExiled = effects.CardTail("moldgraf-monstrosity/after-exile", func(ctx *effects.Context, _ game.TailOutcome) error {
	return moldgrafReturnTwoAtRandom(ctx)
})

// moldgrafReturnTwoAtRandom returns two creature cards chosen at random from
// the controller's graveyard to the battlefield.
func moldgrafReturnTwoAtRandom(ctx *effects.Context) error {
	ids := effects.GraveyardIDs(ctx.Game, ctx.Controller(), func(c game.Card) bool { return c.IsCreature() })
	pick := ctx.Game.ChooseAtRandomForEffect(effects.RandomDraw(ctx), ids, 2)
	if len(pick) == 0 {
		return nil
	}
	return effects.ReturnFromGraveyardTogether{Targets: pick, Controller: ctx.Controller()}.Apply(ctx)
}
