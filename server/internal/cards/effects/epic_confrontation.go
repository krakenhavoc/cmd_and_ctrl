package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Epic Confrontation — Sorcery {1}{G}:
//
//	"Target creature you control gets +1/+2 until end of turn. It fights
//	 target creature you don't control."
//
// The pump resolves first, so the boosted power is the damage dealt in the
// fight (CR 608.2c, 701.14). If either target is gone the fight does not
// happen (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "31123cb5-2dd5-4f2b-a2fd-8f87095972c6",
		Name:         "Epic Confrontation",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			TargetCreature("target creature you don't control", OpponentControls()),
		),
		Purpose: ForTargets(BenefitUntilEOT(0, 1, 2)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			mine, ok := ctx.ClauseTarget(0)
			if !ok || mine.Kind != game.TargetCard {
				return nil
			}
			theirs, ok := ctx.ClauseTarget(1)
			if !ok || theirs.Kind != game.TargetCard {
				return nil
			}
			if err := (BoostUntilEOT{Target: mine.ID, Power: 1, Toughness: 2, Label: "Epic Confrontation — +1/+2 until end of turn"}).Apply(ctx); err != nil {
				return err
			}
			return b10Fight(ctx, mine.ID, theirs.ID)
		},
	})
}
