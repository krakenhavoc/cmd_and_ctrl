package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Power Conduit — Artifact {2}:
//
//	"{T}, Remove a counter from a permanent you control: Choose one —
//	 • Put a charge counter on target artifact.
//	 • Put a +1/+1 counter on target creature."
//
// Scholar of New Horizons' cost (a counter of any kind, off any
// permanent you control, paid at announce, Power Conduit itself
// included) on a modal activated ability (Cankerbloom's shape). The
// mode and its target are announced with the activation, so the
// counter is gone whatever the answer and a target that left makes
// that mode do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2e358f7e-c282-4b92-8255-1436c99cda49",
		Name:         "Power Conduit",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Remove a counter from a permanent you control: Choose one — put a charge counter on target artifact; or put a +1/+1 counter on target creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.Plus(effects.TapCost(), effects.RemoveCountersFrom("", 1, "a permanent you control", effects.Permanent())),
			Modes: effects.ChooseOne(
				effects.ModeDoing("Put a charge counter on target artifact.",
					effects.TargetPermanent("target artifact", effects.Artifact()),
					func(_ *game.StackItem, ctx *effects.Context, occ int) error {
						t, ok := effects.ModeTarget(ctx, occ)
						if !ok {
							return nil
						}
						return effects.AddCounter{Target: t.ID, Kind: game.CounterCharge, N: 1}.Apply(ctx)
					}),
				effects.ModeDoing("Put a +1/+1 counter on target creature.",
					effects.TargetCreature("target creature"),
					effects.PlusOneCounterOnTheModesTarget),
			),
		}},
	})
}
