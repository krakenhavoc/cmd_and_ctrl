package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "14bc9e20-a6ad-4eca-8c5c-5459ce078d9a",
		Name:            "Sigardian Savior",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Answers: game.AnswerValue},
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventETB, AllOf(Self, crestingMosasaurusYouCastIt),
					"Sigardian Savior — return up to two target creature cards with mana value 2 or less from your graveyard to the battlefield",
					func(g *game.Game, item *game.StackItem) error {
						if !b16EnteredFromStack(g, item.SourceCardID) {
							return nil
						}
						ctx := NewContext(g, item)
						return ReturnFromGraveyardTogether{
							Targets:    legalTargetCardIDs(ctx),
							Controller: ctx.Controller(),
						}.Apply(ctx)
					}),
				TargetCardInGraveyard("up to two target creature cards with mana value 2 or less from your graveyard",
					YouOwn(), Creature(), ManaValueLE(2)).WithCount(0, 2)),
		},
	})
}
