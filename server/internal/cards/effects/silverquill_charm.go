package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Silverquill Charm — Instant {W}{B} (Secrets of Strixhaven, #4456):
//
//	"Choose one —
//	 • Put two +1/+1 counters on target creature.
//	 • Exile target creature with power 2 or less.
//	 • Each opponent loses 3 life and you gain 3 life."
//
// A modal spell (ChooseOne) whose bullets carry their own bodies, so each
// reads its own target group. "Power 2 or less" is checked as the spell
// resolves (CR 608.2b), against the creature's current power. The drain is
// two instructions of one mode: each opponent loses the life, then you gain it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d985b12e-2731-4ccf-9f51-90218b20b42a",
		Name:         "Silverquill Charm",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Put two +1/+1 counters on target creature.",
				TargetCreature("target creature"),
				rfCreatureBModeTargetDoes(func(ctx *Context, id uuid.UUID) error {
					return AddCounter{Target: id, Kind: game.CounterPlusOne, N: 2}.Apply(ctx)
				})),
			ModeDoing("Exile target creature with power 2 or less.",
				TargetCreature("target creature with power 2 or less", PowerLE(2)),
				exileTheModesTarget),
			ModeDoing("Each opponent loses 3 life and you gain 3 life.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					if err := (EachOpponentLosesLife{N: 3}).Apply(ctx); err != nil {
						return err
					}
					return GainLife{Player: ctx.Controller(), Amount: 3}.Apply(ctx)
				}),
		),
	})
}
