package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Barishi — Creature — Elemental {2}{G}{G}, 4/3 (Weatherlight):
//
//	"When this creature dies, exile it, then shuffle all creature cards from your
//	 graveyard into your library."
//
// The exile comes first so Barishi itself is not among the creature cards
// shuffled back. The library is shuffled even when no creature card is left
// (the same shape as Finale of Revelation's "shuffle your graveyard into your
// library"). The two instructions are sequential, not conditional.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ffa93a79-7a10-4331-91de-eabbc4fee17e",
		Name:         "Barishi",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Barishi — exile it, then shuffle all creature cards from your graveyard into your library",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneGraveyard {
						return barishiShuffle(ctx)
					}
					return effects.ExileTarget{Target: item.SourceCardID, ThenTail: ctx.ItemTail(barishiExiled)}.Apply(ctx)
				}),
		},
	})
}

// barishiExiled is the clause after the exile. An on-disk key.
var barishiExiled = effects.CardTail("s82-old-b775/barishi-after-exile", func(ctx *effects.Context, _ game.TailOutcome) error {
	return barishiShuffle(ctx)
})

func barishiShuffle(ctx *effects.Context) error {
	you := ctx.Controller()
	for _, id := range effects.GraveyardIDs(ctx.Game, you, func(c game.Card) bool { return c.IsCreature() }) {
		if err := (effects.ReturnFromGraveyard{Target: id, Dest: game.ZoneLibrary}).Apply(ctx); err != nil {
			return err
		}
	}
	return effects.ShuffleLibrary{Player: you}.Apply(ctx)
}
