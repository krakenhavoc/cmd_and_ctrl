package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Insurrection — Sorcery {5}{R}{R}{R}:
//
//	"Untap all creatures and gain control of them until end of turn.
//	 They gain haste until end of turn."
//
// A table-wide Act of Treason (#756): the same three primitives —
// untap, gain control until end of turn, grant haste until end of
// turn — run once per creature on the battlefield, snapshotted before
// anything moves (MatchingBattlefield, the mass.go iterate-the-board
// shape Blasphemous Act's sweep uses). "All creatures" is every
// creature on the board, including the caster's own; gaining control
// of a creature you already control is a harmless no-op layer-2
// effect, exactly as it is in paper.
//
// Printed order is untap, then gain control, then haste; each
// primitive is applied to every creature before the next runs, which
// only matters for a replacement watching one of the three events
// mid-sweep and is otherwise unobservable.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7f7c204f-be1a-47e4-91c6-ba6f906d9012",
		Name:         "Insurrection",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			creatures := MatchingBattlefield(ctx, Creature())
			for _, c := range creatures {
				if err := (UntapTarget{Target: c.InstanceID}).Apply(ctx); err != nil {
					return err
				}
			}
			for _, c := range creatures {
				if err := (GainControl{
					Target:   c.InstanceID,
					Duration: DurationUntilEndOfTurn(ctx),
					Label:    "Insurrection — gain control until end of turn",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			for _, c := range creatures {
				if err := (GrantKeywordUntilEOT{
					Target:   c.InstanceID,
					Keywords: []string{"haste"},
					Label:    "Insurrection — haste until end of turn",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
