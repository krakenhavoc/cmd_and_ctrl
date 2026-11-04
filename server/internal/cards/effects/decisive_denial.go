package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Decisive Denial — Instant {G}{U}:
//
//	"Choose one —
//	 • Target creature you control fights target creature you don't
//	   control. (Each deals damage equal to its power to the other.)
//	 • Counter target noncreature spell unless its controller pays {3}."
//
// The first bullet is Dromoka's Command's fight (two clauses, the fight
// happening only while both are still legal); the second is
// Confusticate and Bebother's counter-unless-pays on a noncreature
// spell (the pay prompt goes to the spell's controller, declining
// counters it).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2d0bd780-b754-44a3-894c-a9c8aa711f05",
		Name:         "Decisive Denial",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Target creature you control fights target creature you don't control.",
				Clauses(
					TargetCreature("target creature you control", YouControl()),
					TargetCreature("target creature you don't control", OpponentControls()),
				),
				FightTheModesTargets),
			ModeDoing("Counter target noncreature spell unless its controller pays {3}.",
				TargetSpell("target noncreature spell", Noncreature()),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return CounterUnlessPaid{
						StackID:  t.ID,
						Cost:     "{3}",
						Question: "Decisive Denial — pay {3} or your spell is countered",
					}.Apply(ctx)
				}),
		),
	})
}
