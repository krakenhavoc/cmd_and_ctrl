package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Savage Smash — Sorcery {1}{R}{G} (Breaking News, #5929):
//
//	"Target creature you control gets +2/+2 until end of turn. It fights
//	 target creature you don't control."
//
// The pump (layer 7c, until end of turn) goes first, then the fight between
// the still-legal picks, so the +2/+2 counts in the damage dealt. A creature
// you control that is no longer legal gets no pump and no fight; if only the
// other target is gone, the pump still happens (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "676c0315-1cb3-4d80-b136-16d88fd612cb",
		Name:         "Savage Smash",
		Completeness: CompletenessFull,
		Purpose: game.Purpose{
			Answers: game.AnswerRemove,
			Targets: game.ForTargets(BenefitUntilEOT(0, 2, 2)),
		},
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			TargetCreature("target creature you don't control", OpponentControls()),
		),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				mine, ok := ctx.ClauseTarget(0)
				if !ok || mine.Kind != game.TargetCard {
					return nil
				}
				return BoostUntilEOT{Target: mine.ID, Power: 2, Toughness: 2, Label: "Savage Smash — +2/+2"}.Apply(ctx)
			}),
			Instruction(fightTargetClauses),
		),
	})
}
