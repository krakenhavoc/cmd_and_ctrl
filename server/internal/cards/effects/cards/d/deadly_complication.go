package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

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
	effects.Register(effects.Spec{
		OracleID:     "c53c75a0-c248-4e76-8d28-f7e53b2dcc49",
		Name:         "Deadly Complication",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseN("Choose one or both", 1, 2,
			effects.ModeDoing("Destroy target creature.",
				effects.TargetCreature("target creature"), effects.DestroyTheModesTarget),
			effects.ModeDoing("Put a +1/+1 counter on target suspected creature you control. You may have it become no longer suspected.",
				effects.TargetCreature("target suspected creature you control", effects.Suspected(), effects.YouControl()), deadlyComplicationCounterAndRelease),
		),
	})
}

// deadlyComplicationCounterAndRelease is the second bullet: the counter,
// then the optional "no longer suspected".
func deadlyComplicationCounterAndRelease(_ *game.StackItem, ctx *effects.Context, occ int) error {
	t, ok := effects.ModeTarget(ctx, occ)
	if !ok {
		return nil
	}
	if err := (effects.AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
		return err
	}
	return effects.MayChoice{
		Question: "Deadly Complication — make it no longer suspected?",
		YesLabel: "No longer suspected",
		NoLabel:  "Keep it suspected",
		Then:     deadlyComplicationUnsuspect,
		Carry:    []uuid.UUID{t.ID},
	}.Apply(ctx)
}

// deadlyComplicationUnsuspect is the "yes": the target (the carried ID) is
// no longer suspected. Keyed (#4291).
var deadlyComplicationUnsuspect = effects.MayYes("confirm/deadly-complication-unsuspect", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	return effects.Unsuspect{Target: effects.CarryAt(r, 0)}.Apply(ctx)
})
