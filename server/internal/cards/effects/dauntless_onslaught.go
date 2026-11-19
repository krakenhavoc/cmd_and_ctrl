package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dauntless Onslaught — Instant {2}{W} (Foundations Jumpstart, #5929):
//
//	"Up to two target creatures each get +2/+2 until end of turn."
//
// One target clause of zero to two creatures; each still legal as the
// spell resolves gets the boost (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ce29f71a-6662-4be4-9915-5a4be87288e0",
		Name:         "Dauntless Onslaught",
		Purpose:      ForTargets(BenefitUntilEOT(0, 2, 2)),
		Completeness: CompletenessFull,
		Targets:      TargetCreature("up to two target creatures").WithCount(0, 2),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (BoostUntilEOT{Target: t.ID, Power: 2, Toughness: 2, Label: "Dauntless Onslaught — +2/+2"}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
