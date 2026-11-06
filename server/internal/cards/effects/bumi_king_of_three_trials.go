package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bumi, King of Three Trials — Legendary Creature — Human Noble Ally
// {5}{G}, 4/4 (Avatar: The Last Airbender, #4459):
//
//	"When Bumi enters, choose up to X, where X is the number of Lesson
//	 cards in your graveyard —
//	 • Put three +1/+1 counters on Bumi.
//	 • Target player scries 3.
//	 • Earthbend 3."
//
// A modal trigger whose count is the board's (ChooseUpToCount, #4459's
// seam beside ChooseUpToX): X is counted as the trigger is put on the
// stack, when its modes are chosen (CR 603.3c), and capped at the three
// bullets, since none can be chosen twice (CR 700.2d, the ruling). With
// no Lesson in the graveyard nothing can be chosen and the trigger is
// removed (CR 700.2b). The chosen bullets run in printed order
// (CR 608.2c): the counters go on Bumi only if it is still on the
// battlefield, the scry is the target player's, and earthbend is the
// keyword action on the bullet's own target land.
//
// No simplification.
func init() {
	enters := WhenThisEnters("Bumi, King of Three Trials — choose up to X",
		func(*game.Game, *game.StackItem) error { return nil })
	enters.Modes = ChooseUpToCount(bumiLessonCards,
		ModeDoing("Put three +1/+1 counters on Bumi.", nil,
			func(_ *game.StackItem, ctx *Context, _ int) error {
				return CounterOnThis{Kind: game.CounterPlusOne, N: 3}.Apply(ctx)
			}),
		ModeDoing("Target player scries 3.", TargetPlayer("target player"),
			func(_ *game.StackItem, ctx *Context, occ int) error {
				t, ok := ModeTarget(ctx, occ)
				if !ok || t.Kind != game.TargetPlayer {
					return nil
				}
				return Scry{Player: t.ID, N: 3}.Apply(ctx)
			}),
		ModeDoing("Earthbend 3.", EarthbendTargets(),
			func(_ *game.StackItem, ctx *Context, occ int) error {
				t, ok := ModeTarget(ctx, occ)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				return Earthbend{Target: t.ID, N: 3}.Apply(ctx)
			}),
	)
	Register(Spec{
		OracleID:     "f7493bcf-5c17-4562-a759-37082ec49ff8",
		Name:         "Bumi, King of Three Trials",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{enters},
	})
}

// bumiLessonCards is X: the Lesson cards in the choosing player's
// graveyard. An on-disk-stable registry key.
var bumiLessonCards = game.RegisterModeCount("bumi-king-of-three-trials/lessons-in-your-graveyard",
	func(g *game.Game, you uuid.UUID) int { return avB4LessonsInGraveyard(g, you) })
