package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Common Goal — Sorcery {1}{G} (Star Trek, #4973):
//
//	"One or two target creatures you control each deal damage equal to
//	 their power to target creature an opponent controls."
//
// Two target clauses, as Terrific Team-Up without its pump: slot 0 one or
// two creatures you control (the damage sources), slot 1 the recipient.
// Each source still a legal target deals damage equal to its power as the
// spell resolves to the recipient, all in one damage event, each with its
// own deathtouch and lifelink. A source gone deals nothing; a recipient
// gone takes nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d4030d30-4c24-458c-87a9-c2f80bb68d27",
		Name:         "Common Goal",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: Clauses(
			TargetCreature("one or two target creatures you control", YouControl()).WithCount(1, 2),
			TargetCreature("target creature an opponent controls", OpponentControls()),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			victim, ok := ctx.ClauseTarget(1)
			if !ok || victim.Kind != game.TargetCard {
				return nil
			}
			var sources []game.TargetRef
			for _, t := range ctx.ClauseTargets(0) {
				if t.Kind == game.TargetCard && ctx.IsTargetLegal(t) {
					sources = append(sources, t)
				}
			}
			return eachDealsDamageEqualToItsPower(ctx, sources, victim.ID)
		},
	})
}
