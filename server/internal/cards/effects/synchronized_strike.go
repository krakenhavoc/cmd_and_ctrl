package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Synchronized Strike — Instant {2}{G}
// (Amonkhet, #5535):
//
//	"Untap up to two target creatures. They each get +2/+2 until end of turn."
//
// Join Forces's text, for each still-legal target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "59d9070c-dc60-420e-aa0e-32cfb97dfe88",
		Name:         "Synchronized Strike",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 2, 2)),
		Targets:      TargetCreature("up to two target creatures").WithCount(0, 2),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, id := range legalTargetCardIDs(ctx) {
				if err := (UntapTarget{Target: id}).Apply(ctx); err != nil {
					return err
				}
				if err := (BoostUntilEOT{Target: id, Power: 2, Toughness: 2, Label: "Synchronized Strike — +2/+2"}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
