package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deadly Complication — Sorcery {1}{B}{R}:
//
//	"Choose one or both —
//	 • Destroy target creature.
//	 • Put a +1/+1 counter on target suspected creature you control.
//	   You may have it become no longer suspected."
//
// Each bullet reads its own target, in printed order (CR 608.2c), so
// one creature can be both: it is destroyed, and the counter finds
// nothing. The second bullet's clause is the Suspected() predicate, so
// the picker offers only suspected creatures you control; the "may" is
// asked after the counter is placed, and a creature that left in
// response gets neither.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c53c75a0-c248-4e76-8d28-f7e53b2dcc49",
		Name:         "Deadly Complication",
		Completeness: CompletenessFull,
		Modes: ChooseN("Choose one or both", 1, 2,
			ModeDoing("Destroy target creature.",
				TargetCreature("target creature"), DestroyTheModesTarget),
			ModeDoing("Put a +1/+1 counter on target suspected creature you control. You may have it become no longer suspected.",
				TargetCreature("target suspected creature you control", Suspected(), YouControl()), deadlyComplicationCounterAndRelease),
		),
	})
}

// deadlyComplicationCounterAndRelease is the second bullet: the counter,
// then the optional "no longer suspected".
func deadlyComplicationCounterAndRelease(_ *game.StackItem, ctx *Context, occ int) error {
	t, ok := ModeTarget(ctx, occ)
	if !ok {
		return nil
	}
	if err := (AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
		return err
	}
	return MayChoice{
		Question: "Deadly Complication — make it no longer suspected?",
		YesLabel: "No longer suspected",
		NoLabel:  "Keep it suspected",
		OnYes: func(ctx *Context) error {
			return Unsuspect{Target: t.ID}.Apply(ctx)
		},
	}.Apply(ctx)
}
