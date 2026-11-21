package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Join Forces — Instant {2}{W} (Dominaria United, #5487):
//
//	"Untap up to two target creatures. They each get +2/+2 until end of
//	 turn."
//
// "Up to two" is a clause that allows zero to two picks. Each target that
// is still legal as the spell resolves is untapped and pumped; one that has
// gone is skipped (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6f8bf849-890e-4fc7-981f-37ff4b4fb844",
		Name:         "Join Forces",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2)),
		Targets:      effects.TargetCreature("up to two target creatures").WithCount(0, 2),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, id := range effects.LegalTargetCardIDs(ctx) {
				if err := (effects.UntapTarget{Target: id}).Apply(ctx); err != nil {
					return err
				}
				if err := (effects.BoostUntilEOT{Target: id, Power: 2, Toughness: 2, Label: "Join Forces — +2/+2"}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
