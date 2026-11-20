package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mischief and Mayhem — Sorcery {4}{G} (Born of the Gods,
// #5559):
//
//	"Up to two target creatures each get +4/+4 until end of turn."
//
// Coordinated Assault's loop without its keyword: each pick still legal
// as the spell resolves gets the bonus (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "831273d4-ced3-4659-b1f6-d5a0aeed3421",
		Name:         "Mischief and Mayhem",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 4, 4)),
		Targets:      TargetCreature("up to two target creatures").WithCount(0, 2),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (BoostUntilEOT{Target: t.ID, Power: 4, Toughness: 4, Label: "Mischief and Mayhem — +4/+4"}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
