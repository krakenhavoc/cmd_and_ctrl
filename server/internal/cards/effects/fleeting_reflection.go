package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fleeting Reflection — Instant {1}{U} (Outlaws of Thunder Junction,
// #4468):
//
//	"Target creature you control gains hexproof until end of turn. Untap
//	 that creature. Until end of turn, it becomes a copy of up to one
//	 other target creature."
//
// Two target clauses (Bionic Blow's shape): a creature you control, and
// up to one other creature anyone controls. The steps run in printed
// order (CR 608.2c) on the first target, and each clause is re-checked on
// its own (CR 608.2b): with the first target gone, nothing happens; with
// no second target, or one that became illegal, the creature still gains
// hexproof and untaps. The copy is Cytoshape's BecomeCopy until end of
// turn (CR 707.2), a layer 1 effect, so the hexproof granted a moment
// earlier in layer 6 stays on the copy.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c915f815-77d7-4e08-b178-82b29a58bb03",
		Name:         "Fleeting Reflection",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 0, 0, "hexproof")),
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			Distinct(TargetCreature("up to one other target creature").WithCount(0, 1)),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			if err := (GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"hexproof"}, Label: "Fleeting Reflection — hexproof"}).Apply(ctx); err != nil {
				return err
			}
			if err := ctx.Game.UntapTargetForEffect(t.ID); err != nil {
				return err
			}
			of, ok := ctx.ClauseTarget(1)
			if !ok || of.Kind != game.TargetCard {
				return nil
			}
			return BecomeCopy{
				Targets: []uuid.UUID{t.ID},
				Of:      of.ID,
				Label:   "Fleeting Reflection — becomes a copy until end of turn",
			}.Apply(ctx)
		},
	})
}
