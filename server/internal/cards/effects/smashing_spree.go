package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Smashing Spree — Instant {1}{R} (Marvel Super Heroes Commander,
// #4455): "Target attacking creature gets +3/+3 and gains trample until
// end of turn." The target clause is AttackingCreature; the body is a
// pump and a keyword grant on the one target, both until end of turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f0371d3d-ca11-432a-bebd-684caad5781f",
		Name:         "Smashing Spree",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target attacking creature", AttackingCreature()),
		Purpose:      ForTargets(BenefitUntilEOT(0, 3, 3, "trample")),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (BoostUntilEOT{Target: id, Power: 3, Toughness: 3, Label: "Smashing Spree — +3/+3"}).Apply(ctx); err != nil {
				return err
			}
			return GrantKeywordUntilEOT{Target: id, Keywords: []string{"trample"}, Label: "Smashing Spree — trample"}.Apply(ctx)
		},
	})
}
