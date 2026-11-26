package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Press the Advantage — Instant {2}{G}{G} (Dragons of Tarkir, #5551):
//
//	"Up to two target creatures each get +2/+2 and gain trample until end of turn."
//
// Coordinated Assault's up-to-two shape: each target still legal as the
// spell resolves gets the pump and the keyword (CR 608.2b).
//
// No simplification.

func init() {
	effects.Register(effects.Spec{
		OracleID:     "7a021af0-d54d-4220-ad01-1cfed9a65e98",
		Name:         "Press the Advantage",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("up to two target creatures").WithCount(0, 2),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2, "trample")),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.BoostUntilEOT{Target: t.ID, Power: 2, Toughness: 2, Label: "Press the Advantage — +2/+2"}).Apply(ctx); err != nil {
					return err
				}
				if err := (effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"trample"}, Label: "Press the Advantage — trample"}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
