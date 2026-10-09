package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vigorbloom Charm — Instant {G}{W} (Reality Fracture, tracker #2795):
//
//	"Choose one —
//	 • Target permanent you control gains hexproof and indestructible
//	   until end of turn.
//	 • You draw a card and gain 3 life.
//	 • Put a +1/+1 counter on target creature you control. Then it
//	   fights target creature an opponent controls."
//
// The third bullet places the counter first and fights from the
// placement's continuation, so the fight reads the grown power even when
// a counter replacement pauses the placement. It fights only while both
// creatures are still legal targets (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "826f17d1-3806-4579-8b1b-2082b7b03d69",
		Name:         "Vigorbloom Charm",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Target permanent you control gains hexproof and indestructible until end of turn.",
				TargetPermanent("target permanent you control", YouControl()),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"hexproof", "indestructible"}, Label: "Vigorbloom Charm"}.Apply(ctx)
				}),
			ModeDoing("You draw a card and gain 3 life.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					if err := (DrawCards{Player: ctx.Controller(), N: 1}).Apply(ctx); err != nil {
						return err
					}
					return GainLife{Player: ctx.Controller(), Amount: 3}.Apply(ctx)
				}),
			ModeDoing("Put a +1/+1 counter on target creature you control. Then it fights target creature an opponent controls.",
				Clauses(
					TargetCreature("target creature you control", YouControl()),
					TargetCreature("target creature an opponent controls", OpponentControls()),
				),
				func(item *game.StackItem, ctx *Context, occ int) error {
					mine, ok := ModeClauseTarget(ctx, occ, 0)
					if !ok {
						return nil
					}
					theirs, hasTheirs := ModeClauseTarget(ctx, occ, 1)
					return ctx.Game.AddCounterThenForEffect(mine.ID, game.CounterPlusOne, 1, func(g *game.Game, _ int) error {
						if !hasTheirs {
							return nil
						}
						return b10Fight(NewContext(g, item), mine.ID, theirs.ID)
					})
				}),
		),
	})
}
