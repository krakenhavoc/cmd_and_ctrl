package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chelonian Tackle — Sorcery {2}{G} (Secrets of Strixhaven, #4456):
//
//	"Target creature you control gets +0/+10 until end of turn. Then it
//	 fights up to one target creature an opponent controls. (Each deals
//	 damage equal to its power to the other.)"
//
// Two target clauses: your creature (slot 0) and up to one creature an
// opponent controls (slot 1). The pump comes first and is its own step, so
// the fight sees the +0/+10 (it matters for what the fighter survives, not
// for the damage it deals). The fight is skipped, with the pump kept, when
// the second pick was left empty or either creature is no longer a legal
// target (CR 701.14b, CR 608.2b); a creature that has left the battlefield
// fights nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8c1c519a-6edd-4204-b825-b0feca5afa42",
		Name:         "Chelonian Tackle",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TargetCreature("up to one target creature an opponent controls", effects.OpponentControls()).WithCount(0, 1),
		),
		Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 10)),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				t, ok := ctx.ClauseTarget(0)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				return effects.BoostUntilEOT{
					Target:    t.ID,
					Toughness: 10,
					Label:     "Chelonian Tackle — +0/+10 until end of turn",
				}.Apply(ctx)
			}),
			effects.Instruction(effects.FightTargetClauses),
		),
	})
}
