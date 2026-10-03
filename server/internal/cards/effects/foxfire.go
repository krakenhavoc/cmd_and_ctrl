package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Foxfire — Instant {2}{G}:
//
//	"Untap target attacking creature. Prevent all combat damage that would
//	 be dealt to and dealt by that creature this turn.
//	 Draw a card at the beginning of the next turn's upkeep."
//
// ADR 0108 §7, Delivery PR 7 (#1904): Maze of Ith's effect as a spell,
// then Gravebind's delayed draw at the next upkeep, whoever's turn it is
// (CR 603.7). The draw is scheduled even if the creature has gone: the
// spell resolves as long as one target is legal, and with one target an
// illegal one means it doesn't resolve at all (CR 608.2b).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "6c270e6a-12a7-4323-87ba-6d770986fb48",
		Name:         "Foxfire",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target attacking creature", AttackingCreature()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := untapThenToAndBy(ctx); err != nil {
				return err
			}
			return ScheduleDelayedTrigger{
				At:    game.StepUpkeep,
				Label: "Foxfire — draw a card",
				Body:  drawOneBody,
			}.Apply(ctx)
		},
	})
}
