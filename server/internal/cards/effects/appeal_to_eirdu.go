package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Appeal to Eirdu — Instant {3}{W} (Lorwyn Eclipsed, #4458):
//
//	"Convoke (Your creatures can help cast this spell. Each creature you
//	 tap while casting this spell pays for {1} or one mana of that
//	 creature's color.)
//	 One or two target creatures each get +2/+1 until end of turn."
//
// Heroic Teamwork's pump with convoke in place of teamwork: one clause of
// one or two creatures (WithCount(1, 2), CR 601.2c), and each target still
// legal as the spell resolves gets +2/+1 (CR 608.2b). Convoke is
// Harmonized Crescendo's TapCost, so a creature tapped to pay for it can
// still be one of the targets.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cea09261-41a2-43e5-bf43-c4097e6615c6",
		Name:         "Appeal to Eirdu",
		Completeness: CompletenessFull,
		TapCost:      Convoke(),
		Purpose: game.Purpose{
			Answers: game.AnswerPump,
			Targets: game.ForTargets(BenefitUntilEOT(0, 2, 1)),
		},
		Targets: TargetCreature("one or two target creatures").WithCount(1, 2),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if err := (BoostUntilEOT{Target: t.ID, Power: 2, Toughness: 1, Label: "Appeal to Eirdu — +2/+1"}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
