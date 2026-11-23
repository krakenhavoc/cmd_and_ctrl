package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skulduggery — Instant {B} (Outlaws of Thunder Junction, #4468):
//
//	"Until end of turn, target creature you control gets +1/+1 and target
//	 creature an opponent controls gets -1/-1."
//
// Dissection Practice's two creature clauses, each required and each
// bound to a side of the table: slot 0 a creature you control, slot 1
// one an opponent controls. Each is re-judged as the spell resolves
// (CR 608.2b), so one still-legal target is still changed when the other
// has gone. Targeting the opponent's creature is a crime (CR 700.13).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9100530e-2bd0-487e-a4eb-13dabd55b678",
		Name:         "Skulduggery",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
		),
		Purpose: effects.ForTargets(
			effects.BenefitUntilEOT(0, 1, 1),
			effects.HarmUntilEOT(1, -1, -1),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if c, ok := ctx.ClauseTarget(0); ok && c.Kind == game.TargetCard {
				if err := (effects.BoostUntilEOT{Target: c.ID, Power: 1, Toughness: 1, Label: "Skulduggery — +1/+1 until end of turn"}).Apply(ctx); err != nil {
					return err
				}
			}
			if c, ok := ctx.ClauseTarget(1); ok && c.Kind == game.TargetCard {
				return effects.BoostUntilEOT{Target: c.ID, Power: -1, Toughness: -1, Label: "Skulduggery — -1/-1 until end of turn"}.Apply(ctx)
			}
			return nil
		},
	})
}
