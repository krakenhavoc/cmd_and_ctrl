package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "9100530e-2bd0-487e-a4eb-13dabd55b678",
		Name:         "Skulduggery",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			TargetCreature("target creature an opponent controls", OpponentControls()),
		),
		Purpose: ForTargets(
			BenefitUntilEOT(0, 1, 1),
			HarmUntilEOT(1, -1, -1),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if c, ok := ctx.ClauseTarget(0); ok && c.Kind == game.TargetCard {
				if err := (BoostUntilEOT{Target: c.ID, Power: 1, Toughness: 1, Label: "Skulduggery — +1/+1 until end of turn"}).Apply(ctx); err != nil {
					return err
				}
			}
			if c, ok := ctx.ClauseTarget(1); ok && c.Kind == game.TargetCard {
				return BoostUntilEOT{Target: c.ID, Power: -1, Toughness: -1, Label: "Skulduggery — -1/-1 until end of turn"}.Apply(ctx)
			}
			return nil
		},
	})
}
