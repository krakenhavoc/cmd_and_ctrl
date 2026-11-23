package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tam's Resistance — Sorcery {1}{G/U} (Reality Fracture):
//
//	"Put a +1/+1 counter on up to one target creature. It gains
//	 vigilance until end of turn.
//	 Empower Jace 4."
//
// ADR 0139. "Up to one" may be zero targets, and then only the Jace is
// empowered.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7500b7dc-588d-46bd-acc5-885dcd464668",
		Name:         "Tam's Resistance",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("up to one target creature").WithCount(0, 1),
		// Vigilance and Empower Jace wait for a prompt the counter asks (a CR 616 order;
		// CR 608.2c, #4155).
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				for _, t := range ctx.LegalTargets() {
					if err := (effects.AddCounter{Target: t.ID, Kind: "+1/+1", N: 1}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				for _, t := range ctx.LegalTargets() {
					if err := (effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"vigilance"}, Label: "Tam's Resistance — vigilance until end of turn"}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			}),
			effects.EmpowerJace{N: 4},
		),
	})
}
