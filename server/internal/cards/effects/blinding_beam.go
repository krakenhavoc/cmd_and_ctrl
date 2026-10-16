package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blinding Beam — Instant {2}{W}:
//
//	"Choose one —
//	 • Tap two target creatures.
//	 • Creatures don't untap during target player's next untap step.
//	 Entwine {1} (Choose both if you pay the entwine cost.)"
//
// The second bullet is #3995's player-wide skip: it names no creature,
// so it reaches every creature that would untap in that player's next
// untap step, including one they cast or take after Blinding Beam
// resolved (CR 611.2c). The printed "Creatures" names no controller, so
// a creature another player controls that would untap in that step (a
// Seedborn Muse's) stays tapped too. Entwined (CR 702.42) the tap
// happens first (CR 702.42b), so the two tapped creatures sit out the
// step when their controller is the target player.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "0766cf02-ea55-4271-8c73-213c93b0baac",
		Name:          "Blinding Beam",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Entwine("{1}")},
		Modes: Entwining(ChooseOne(
			ModeDoing("Tap two target creatures.",
				TargetCreature("two target creatures").WithCount(2, 2),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					for _, t := range ctx.ModeTargets(occ) {
						if t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
							continue
						}
						if err := (TapTarget{Target: t.ID}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
			ModeDoing("Creatures don't untap during target player's next untap step.",
				TargetPlayer("target player"),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					return PermanentsDontUntapNextStep{
						Player:        t.ID,
						Types:         []string{"creature"},
						AnyController: true,
						Label:         "Blinding Beam — creatures don't untap during that player's next untap step",
					}.Apply(ctx)
				}),
		)),
	})
}
