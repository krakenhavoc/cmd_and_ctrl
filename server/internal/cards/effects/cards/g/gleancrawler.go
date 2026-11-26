package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gleancrawler — Creature — Insect Horror {3}{B/G}{B/G}{B/G}, 6/6 (Duel Decks:
// Izzet vs. Golgari):
//
//	"({B/G} can be paid with either {B} or {G}.)
//	 Trample
//	 At the beginning of your end step, return to your hand all creature cards
//	 in your graveyard that were put there from the battlefield this turn."
//
// Faith's Reward's set, read as the ability resolves: a creature card counts
// when its newest arrival in your graveyard this turn came from the battlefield
// (a card discarded or milled is not one, and a token is not a card). They all
// return to hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "bb7389fa-ab01-4ca2-bd18-03cf5b4cedf6",
		Name:            "Gleancrawler",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("Gleancrawler — return to your hand all creature cards in your graveyard that were put there from the battlefield this turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, id := range effects.GraveyardCardIDs(ctx, ctx.Controller(), func(c game.Card) bool {
						return c.IsCreature() && effects.PutIntoGraveyardFromBattlefieldThisTurn(g, c.InstanceID)
					}) {
						if err := (effects.ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
		},
	})
}
