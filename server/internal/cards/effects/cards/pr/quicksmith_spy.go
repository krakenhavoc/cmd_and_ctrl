package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Quicksmith Spy — Creature — Human Artificer {3}{U}, 2/3 (Aether Revolt, S82
// older sets, #5536):
//
//	"When this creature enters, target artifact you control gains "{T}: Draw
//	 a card" for as long as you control this creature."
//
// The granted ability is an ADR 0093 bundle given by a duration grant that
// lasts while the same Spy stays under its controller's control (CR 611.2b):
// one that has left and come back before the trigger resolves is a new
// object and grants nothing (CR 400.7). The artifact is the host of the
// ability, so {T} taps it and a summoning-sick creature artifact cannot use
// it (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "54250029-1bd7-4caf-a839-9497607ce725",
		Name:         "Quicksmith Spy",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key:  quicksmithSpyDrawGrant,
			Text: "{T}: Draw a card.",
			Activated: []effects.ActivatedAbility{{
				Label:   "{T}: Draw a card.",
				Purpose: game.Purpose{Draws: 1, Answers: game.AnswerValue},
				Cost:    effects.TapCost(),
				Effect:  effects.Do(effects.DrawCards{N: 1}),
			}},
		}},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Quicksmith Spy — target artifact you control gains \"{T}: Draw a card\" for as long as you control this creature",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					d, ok := effects.DurationWhileYouControlSource(ctx, ctx.Source(), ctx.Controller())
					if !ok {
						return nil
					}
					return effects.GrantAbilitiesFor{
						Target:   id,
						Keys:     []string{quicksmithSpyDrawGrant},
						Duration: d,
						Label:    "Quicksmith Spy — {T}: Draw a card",
					}.Apply(ctx)
				}),
				effects.TargetPermanent("target artifact you control", effects.Artifact(), effects.YouControl())),
		},
	})
}

// quicksmithSpyDrawGrant is the granted bundle's key, an on-disk identity:
// never renamed.
const quicksmithSpyDrawGrant = "quicksmith-spy/draw"
