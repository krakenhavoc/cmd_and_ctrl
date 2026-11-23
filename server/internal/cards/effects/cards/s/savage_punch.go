package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Savage Punch — Sorcery {1}{G} (Jumpstart 2022, #5948):
//
//	"Target creature you control fights target creature you don't
//	 control.
//	 Ferocious — The creature you control gets +2/+2 until end of turn
//	 before it fights if you control a creature with power 4 or
//	 greater."
//
// Chelonian Tackle's pump-then-fight over the two target clauses; the
// ferocious check is made as the spell resolves, before the pump and so
// before the fight.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3ca58e01-21c6-4314-ab3d-280ba6f9f509",
		Name:         "Savage Punch",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TargetCreature("target creature you don't control", effects.OpponentControls()),
		),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				t, ok := ctx.ClauseTarget(0)
				if !ok || t.Kind != game.TargetCard || !effects.FerociousYouControlPowerFourOrGreater(ctx) {
					return nil
				}
				return effects.BoostUntilEOT{Target: t.ID, Power: 2, Toughness: 2, Label: "Savage Punch — ferocious +2/+2"}.Apply(ctx)
			}),
			effects.Instruction(effects.FightTargetClauses),
		),
	})
}
