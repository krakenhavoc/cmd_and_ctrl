package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sigardian Savior — Creature — Angel {3}{W}{W}, 3/3 (Midnight Hunt
// Commander, #5497):
//
//	"Flying
//	 When this creature enters, if you cast it, return up to two target
//	 creature cards with mana value 2 or less from your graveyard to the
//	 battlefield."
//
// Bortuk Bonerattle's "if you cast it" (the permanent came from the
// stack, b16EnteredFromStack), checked as the ability triggers and again
// as it resolves (CR 603.4). The targets are Fight On's "up to two" in
// your graveyard, narrowed to creature cards with mana value 2 or less
// (Balduvian Atrocity's filter). The cards that are still legal targets
// enter together as one entry (CR 603.6a); with no target chosen, or none
// still legal, nothing happens (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "14bc9e20-a6ad-4eca-8c5c-5459ce078d9a",
		Name:            "Sigardian Savior",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Answers: game.AnswerValue},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventETB, effects.AllOf(effects.Self, effects.CrestingMosasaurusYouCastIt),
					"Sigardian Savior — return up to two target creature cards with mana value 2 or less from your graveyard to the battlefield",
					func(g *game.Game, item *game.StackItem) error {
						if !effects.B16EnteredFromStack(g, item.SourceCardID) {
							return nil
						}
						ctx := effects.NewContext(g, item)
						return effects.ReturnFromGraveyardTogether{
							Targets:    effects.LegalTargetCardIDs(ctx),
							Controller: ctx.Controller(),
						}.Apply(ctx)
					}),
				effects.TargetCardInGraveyard("up to two target creature cards with mana value 2 or less from your graveyard",
					effects.YouOwn(), effects.Creature(), effects.ManaValueLE(2)).WithCount(0, 2)),
		},
	})
}
