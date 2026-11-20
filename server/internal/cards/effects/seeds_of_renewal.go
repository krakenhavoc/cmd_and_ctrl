package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Seeds of Renewal — Sorcery {6}{G} (Commander 2016, #5537):
//
//	"Undaunted (This spell costs {1} less to cast for each opponent.)
//	 Return up to two target cards from your graveyard to your hand.
//	 Exile Seeds of Renewal."
//
// Curtains' Call's undaunted (CR 702.125a) on Retrieve's shape: the
// still-legal picks return to hand (a card that left the graveyard in
// response stays where it went, CR 608.2b), then the spell exiles itself
// instead of going to the graveyard. A spell countered, or one whose
// every target became illegal, never resolved and is not exiled.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "48d63b86-8ddd-4306-960f-98490d1d80b9",
		Name:         "Seeds of Renewal",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets:      TargetCardInGraveyard("up to two target cards from your graveyard", YouOwn()).WithCount(0, 2),
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(OpponentsOfCaster(), "Undaunted"),
		},
		OnResolve: ResolveSteps(
			Instruction(returnLegalGraveyardTargetsToHand),
			Instruction(func(ctx *Context) error {
				// "Exile Seeds of Renewal." The spell moves itself, so the
				// resolution frame does not put it into the graveyard (#489).
				return ctx.Game.ExileCardForEffect(ctx.Item.SourceCardID)
			}),
		),
	})
}
