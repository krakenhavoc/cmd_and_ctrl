package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// You Come to the Gnoll Camp — Instant {1}{R} (Adventures in the Forgotten
// Realms, #5499):
//
//	"Choose one —
//	 • Intimidate Them — Up to two target creatures can't block this turn.
//	 • Fend Them Off — Target creature gets +3/+1 until end of turn."
//
// Beat a Path's block restriction over up to two targets (a bullet with no
// target at all still resolves) and Flourishing Strike's pump on the
// bullet's own target. Each target is judged as the spell resolves
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e0945337-1833-4545-a83f-fca673e7f0d0",
		Name:         "You Come to the Gnoll Camp",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Intimidate Them — Up to two target creatures can't block this turn.",
				effects.TargetCreature("up to two target creatures").WithCount(0, 2),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					for _, t := range ctx.ModeTargets(occ) {
						if t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
							continue
						}
						if err := (effects.RestrictUntilEOT{Target: t.ID, Restrictions: game.CantBlock, Label: "You Come to the Gnoll Camp — can't block"}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}), effects.ForTargets(effects.CantBlockUntilEOT(0, 0, 0))),
			effects.ModeWithPurpose(effects.ModeDoing("Fend Them Off — Target creature gets +3/+1 until end of turn.",
				effects.TargetCreature("target creature"),
				effects.BoostTheModesTarget(3, 1, "You Come to the Gnoll Camp — Fend Them Off")), effects.ForTargets(effects.BenefitUntilEOT(0, 3, 1))),
		),
	})
}
