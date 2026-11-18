package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "2e358f7e-c282-4b92-8255-1436c99cda49",
		Name:         "Power Conduit",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}, Remove a counter from a permanent you control: Choose one — put a charge counter on target artifact; or put a +1/+1 counter on target creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    Plus(TapCost(), RemoveCountersFrom("", 1, "a permanent you control", Permanent())),
			Modes: ChooseOne(
				ModeDoing("Put a charge counter on target artifact.",
					TargetPermanent("target artifact", Artifact()),
					func(_ *game.StackItem, ctx *Context, occ int) error {
						t, ok := ModeTarget(ctx, occ)
						if !ok {
							return nil
						}
						return AddCounter{Target: t.ID, Kind: game.CounterCharge, N: 1}.Apply(ctx)
					}),
				ModeDoing("Put a +1/+1 counter on target creature.",
					TargetCreature("target creature"),
					PlusOneCounterOnTheModesTarget),
			),
		}},
	})
}
