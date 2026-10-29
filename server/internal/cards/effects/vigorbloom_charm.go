package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

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
			ModeWithPurpose(ModeDoing("Target permanent you control gains hexproof and indestructible until end of turn.",
				TargetPermanent("target permanent you control", YouControl()),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"hexproof", "indestructible"}, Label: "Vigorbloom Charm"}.Apply(ctx)
				}), ForTargets(BenefitUntilEOT(0, 0, 0, "hexproof", "indestructible"))),
			// The life gain waits for the draw's prompt (a dredge offer; CR 608.2c, #4155).
			ModeSteps("You draw a card and gain 3 life.", nil,
				DrawCards{N: 1},
				GainLife{Amount: 3}),
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
					then := ctx.ItemTail(vigorbloomCharmFight)
					then.Carry = []uuid.UUID{mine.ID}
					if hasTheirs {
						then.Carry = append(then.Carry, theirs.ID)
					}
					return ctx.Game.AddCounterByKeyedForEffect(uuid.Nil, mine.ID, game.CounterPlusOne, 1, then)
				}),
		),
	})
}

// vigorbloomCharmFight is "Then it fights target creature an opponent
// controls" once the counter has landed; Carry is the creature you control,
// then the other if there still is one. An on-disk key: never renamed, never
// reused.
var vigorbloomCharmFight = CardTail("vigorbloom-charm/fight", func(ctx *Context, o game.TailOutcome) error {
	if len(o.Tail.Carry) < 2 {
		return nil
	}
	return b10Fight(ctx, o.Tail.Carry[0], o.Tail.Carry[1])
})
