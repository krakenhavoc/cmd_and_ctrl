package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Coordinated Assault — Instant {R} (Ravnica: Clue Edition, #5929):
//
//	"Up to two target creatures each get +1/+0 and gain first strike until
//	 end of turn."
//
// One target clause with a count of 0 to 2; each creature still legal as
// the spell resolves gets the boost and the keyword (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ba14843d-6b7f-4e3d-b5ba-fc573fa8bf50",
		Name:         "Coordinated Assault",
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 1, 0, "first strike")),
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("up to two target creatures").WithCount(0, 2),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.BoostUntilEOT{Target: t.ID, Power: 1, Label: "Coordinated Assault — +1/+0"}).Apply(ctx); err != nil {
					return err
				}
				if err := (effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"first strike"}, Label: "Coordinated Assault — first strike"}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
