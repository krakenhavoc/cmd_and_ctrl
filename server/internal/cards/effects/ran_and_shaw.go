package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ran and Shaw — Legendary Creature — Dragon {3}{R}{R}, 4/4:
//
//	"Flying, firebending 2
//	 When Ran and Shaw enter, if you cast them and there are three or
//	 more Dragon and/or Lesson cards in your graveyard, create a token
//	 that's a copy of Ran and Shaw, except it's not legendary.
//	 {3}{R}: Dragons you control get +2/+0 until end of turn."
//
// Firebending 2 (CR 702.189a) adds {R}{R} that lasts until end of
// combat, as Avatar Roku's does.
//
// The enters trigger is an intervening "if" (CR 603.4): it triggers
// only for a cast Ran and Shaw with three such cards already in the
// graveyard, and checks the graveyard again as it resolves. Having been
// cast cannot stop being true, so only the count is read twice. The
// copy is made from the copiable values, with last-known information
// if Ran and Shaw have left (CR 608.2h), and the "not legendary" is
// part of the copy (CR 707.9b). The token was not cast, so its own
// enters trigger does nothing.
//
// The pump locks its set as it resolves (CR 611.2c).
func init() {
	const label = "Ran and Shaw — create a token that's a copy of Ran and Shaw, except it's not legendary"
	Register(Spec{
		OracleID:        "a728685f-8670-4db2-ae02-3cf74eb3c402",
		Name:            "Ran and Shaw",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Ran and Shaw — firebending 2", Do(firebending(2))),
			On(game.EventETB, AllOf(Self, geologicalAppraiserWasCast, ranAndShawGraveyardHasThree), label,
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if avB12DragonOrLessonCardsInGraveyard(g, item.Controller, ctx.TextSubtype("Dragon")) < 3 {
						return nil
					}
					return CreateTokenCopy{
						Controller: item.Controller,
						Copy:       item.SourceCardID,
						N:          1,
						Except: func(t *game.Card) {
							t.TypeLine = removeLegendaryFromTypeLine(t.TypeLine)
						},
					}.Apply(ctx)
				}),
		},
		Activated: []ActivatedAbility{{
			Label:   "{3}{R}: Dragons you control get +2/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{3}{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return BoostUntilEOT{
					Match: And(OfCreatureType("Dragon"), YouControl()),
					Power: 2,
					Label: "Ran and Shaw — Dragons you control get +2/+0",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}

// ranAndShawGraveyardHasThree is the trigger-time half of the
// intervening "if": three or more Dragon and/or Lesson cards in the
// graveyard of the entering permanent's controller.
func ranAndShawGraveyardHasThree(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return source != nil && avB12DragonOrLessonCardsInGraveyard(g, source.Controller, source.TextSubtype("Dragon")) >= 3
}
