package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Caught Red-Handed — Instant {4}{R}:
//
//	"This spell can't be countered. (This includes by the ward
//	 ability.)
//	 Gain control of target creature until end of turn. Untap that
//	 creature. It gains haste until end of turn. Suspect it. (It has
//	 menace and can't block.)"
//
// Act of Treason's three clauses in printed order, then the suspect.
// The designation is not tied to the theft: it is a status of the
// permanent, so the creature goes back to its owner still suspected
// when the turn ends (and the same turn it can't block, menace or not).
//
// "Can't be countered" is Spec.CantBeCountered, as Abrupt Decay.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "484b3586-c0ed-4793-a3c0-e99f265e092d",
		Name:            "Caught Red-Handed",
		Completeness:    CompletenessFull,
		CantBeCountered: true,
		Targets:         TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			if err := (GainControl{
				Target:   target,
				Duration: DurationUntilEndOfTurn(ctx),
				Label:    "Caught Red-Handed — gain control until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			if err := (UntapTarget{Target: target}).Apply(ctx); err != nil {
				return err
			}
			if err := (GrantKeywordUntilEOT{
				Target:   target,
				Keywords: []string{"haste"},
				Label:    "Caught Red-Handed — haste until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			return Suspect{Target: target}.Apply(ctx)
		},
	})
}
