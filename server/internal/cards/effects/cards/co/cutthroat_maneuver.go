package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cutthroat Maneuver — Instant {3}{B} (Theros, S82 OLD b391, #5561):
//
//	"Up to two target creatures each get +1/+1 and gain lifelink until end
//	 of turn."
//
// Triton Tactics' up-to-two clause: each target still legal as the spell
// resolves gets the pump and the keyword; with none legal nothing happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "28a24aeb-1ede-4331-adae-6b2dfcd3e4da",
		Name:         "Cutthroat Maneuver",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("up to two target creatures").WithCount(0, 2),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1, "lifelink")),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.BoostUntilEOT{Target: t.ID, Power: 1, Toughness: 1, Label: "Cutthroat Maneuver — +1/+1"}).Apply(ctx); err != nil {
					return err
				}
				if err := (effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"lifelink"}, Label: "Cutthroat Maneuver — lifelink"}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
