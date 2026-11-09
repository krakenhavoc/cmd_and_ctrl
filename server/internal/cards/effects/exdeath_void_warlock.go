package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// exdeathOracle is shared by both faces: the front registers under it,
// the back under exdeathOracle + "#1".
const exdeathOracle = "98c9d798-8c48-4a2d-a076-ce17d31df6f5"

// Exdeath, Void Warlock // Neo Exdeath, Dimension's End — a transforming
// legend (Final Fantasy, #4462):
//
//	Exdeath, Void Warlock, {1}{B}{G} Legendary Creature — Spirit Warlock,
//	3/3
//	  "When Exdeath enters, you gain 3 life.
//	   At the beginning of your end step, if there are six or more
//	   permanent cards in your graveyard, transform Exdeath."
//
//	Neo Exdeath, Dimension's End, Legendary Creature — Spirit Avatar, */3
//	  "Trample
//	   Neo Exdeath's power is equal to the number of permanent cards in
//	   your graveyard."
//
// The enters trigger is a gain-life step. The end-step trigger is an
// intervening "if" (CR 603.4): checked as the end step begins and again
// as the ability resolves (the ruling), over the permanent cards in its
// controller's graveyard (CR 110.4), and it transforms Exdeath in place
// (TransformThis does nothing if it has already turned over, CR 701.27f).
// The back face's power is a layer 7a characteristic-defining ability
// (CR 604.3) over the same count, read on every recompute; a card
// entering or leaving the graveyard drops the layer cache.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     exdeathOracle,
		Name:         "Exdeath, Void Warlock",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Exdeath, Void Warlock — you gain 3 life", Do(GainLife{Amount: 3})),
			InterveningIf(exdeathSixPermanentCards{}, On(game.EventBeginEndStep,
				AllOf(ByYou, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return permanentCardsInGraveyard(g, source.Controller) >= 6
				}),
				"Exdeath, Void Warlock — transform Exdeath", Do(TransformThis{}))),
		},
	})
	Register(Spec{
		OracleID:        exdeathOracle + "#1",
		Name:            "Neo Exdeath, Dimension's End",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static: []game.StaticAbility{{
			Label:     "Neo Exdeath's power is equal to the number of permanent cards in your graveyard.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power = permanentCardsInGraveyard(g, source.Controller)
			},
		}},
	})
}

// exdeathSixPermanentCards is the intervening "if there are six or more
// permanent cards in your graveyard".
type exdeathSixPermanentCards struct{}

func (exdeathSixPermanentCards) Holds(ctx *Context) bool {
	return permanentCardsInGraveyard(ctx.Game, ctx.Controller()) >= 6
}

func (exdeathSixPermanentCards) condition() (game.FootprintStep, bool) {
	return condStep(game.FootprintCondBoard, 0), true
}
