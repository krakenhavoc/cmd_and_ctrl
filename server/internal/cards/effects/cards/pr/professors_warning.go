package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Professor's Warning — Instant {B} (Strixhaven: School of Mages, #5502):
//
//	"Choose one —
//	 • Put a +1/+1 counter on target creature.
//	 • Target creature gains indestructible until end of turn."
//
// Each bullet has its own target (CR 700.2): the counter is the shared
// PlusOneCountersOnTheModesTarget and the grant is Battle-Rage Blessing's
// keyword grant, indestructible only.
//
// No simplification.
func init() {
	protect := effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "indestructible"))
	protect.Answers = game.AnswerProtect
	effects.Register(effects.Spec{
		OracleID:     "e3c8e552-d315-486d-a176-813b3a300174",
		Name:         "Professor's Warning",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Put a +1/+1 counter on target creature.",
				effects.TargetCreature("target creature"),
				effects.PlusOneCountersOnTheModesTarget(1)),
			effects.ModeWithPurpose(effects.ModeDoing("Target creature gains indestructible until end of turn.",
				effects.TargetCreature("target creature"),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return effects.GrantKeywordUntilEOT{
						Target:   t.ID,
						Keywords: []string{"indestructible"},
						Label:    "Professor's Warning — indestructible",
					}.Apply(ctx)
				}),
				protect),
		),
	})
}
