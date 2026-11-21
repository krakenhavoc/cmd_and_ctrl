package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tifa's Limit Break — Instant {G} (#3777):
//
//	"Tiered (Choose one additional cost.)
//	 • Somersault — {0} — Target creature gets +2/+2 until end of turn.
//	 • Meteor Strikes — {2} — Double target creature's power and
//	   toughness until end of turn.
//	 • Final Heaven — {6}{G} — Triple target creature's power and
//	   toughness until end of turn."
//
// Tiered proof card (CR 702.183a). Doubling is CR 701.10's +X/+Y read
// as the bullet resolves (doubling.go); tripling is the same reading
// added twice, +2X/+2Y, so a later pump is not tripled.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d0d90dd8-0642-489f-9afc-11c307f258f2",
		Name:         "Tifa's Limit Break",
		Completeness: effects.CompletenessFull,
		Modes: effects.Tiered(
			effects.ModeWithPurpose(effects.SpreeModeDoing("Somersault — Target creature gets +2/+2 until end of turn.", "{0}",
				effects.TargetCreature("target creature"),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return effects.BoostUntilEOT{Target: t.ID, Power: 2, Toughness: 2, Label: "Tifa's Limit Break — Somersault"}.Apply(ctx)
				}), effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2))),
			effects.ModeWithPurpose(effects.SpreeModeDoing("Meteor Strikes — Double target creature's power and toughness until end of turn.", "{2}",
				effects.TargetCreature("target creature"), tifaMultiply(2, "Tifa's Limit Break — Meteor Strikes")), effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0))),
			effects.ModeWithPurpose(effects.SpreeModeDoing("Final Heaven — Triple target creature's power and toughness until end of turn.", "{6}{G}",
				effects.TargetCreature("target creature"), tifaMultiply(3, "Tifa's Limit Break — Final Heaven")), effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0))),
		),
	})
}

// tifaMultiply is "multiply target creature's power and toughness by
// k until end of turn": it gets (k-1) times its power and toughness as
// the bullet resolves.
func tifaMultiply(k int, label string) func(*game.StackItem, *effects.Context, int) error {
	return func(_ *game.StackItem, ctx *effects.Context, occ int) error {
		t, ok := effects.ModeTarget(ctx, occ)
		if !ok {
			return nil
		}
		dp, dt, ok := effects.DoubledBy(ctx.Game, t.ID, true)
		if !ok {
			return nil
		}
		return effects.BoostUntilEOT{Target: t.ID, Power: (k - 1) * dp, Toughness: (k - 1) * dt, Label: label}.Apply(ctx)
	}
}
