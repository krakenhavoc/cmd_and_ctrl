package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Satyr Piper — Creature — Satyr Rogue {2}{G}, 2/1 (Theros):
//
//	"{3}{G}: Target creature must be blocked this turn if able."
//
// The requirement is a data record pinned to the target until end of turn
// (Irresistible Prey's BlockRequirementUntilEOT).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d15eed8e-6c2c-4427-9f72-a46e90c562b9",
		Name:         "Satyr Piper",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{G}: Target creature must be blocked this turn if able.",
			Cost:    effects.ManaCost("{3}{G}"),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					if err := (effects.BlockRequirementUntilEOT{
						Target: t.ID,
						Kind:   game.BlockRequirementMustBeBlocked,
						Label:  "Satyr Piper — must be blocked this turn",
					}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
