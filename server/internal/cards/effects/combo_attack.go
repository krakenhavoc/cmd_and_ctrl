package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Combo Attack — Sorcery {2}{G} (Battlebond, #5529):
//
//	"Two target creatures your team controls each deal damage equal to
//	 their power to target creature."
//
// "Your team" is "you" (CR 102.4). Tandem Takedown's damage without the
// boost: two creatures you control, then a creature to hit. The victim is
// its own instance of "target", so it may be one of the two (CR 115.3), and
// then that creature deals damage to itself. Each of the two still legal
// deals damage equal to its power as the spell resolves, together as one
// instance (eachDealsDamageEqualToItsPower); with the victim gone, nothing
// is dealt.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "037eccdb-21c3-4048-af4f-b25dcbe99134",
		Name:         "Combo Attack",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCreature("two target creatures your team controls", YouControl()).WithCount(2, 2),
			TargetCreature("target creature"),
		),
		Purpose: game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			var sources []game.TargetRef
			for _, t := range ctx.ClauseTargets(0) {
				if t.Kind == game.TargetCard && ctx.IsTargetLegal(t) {
					sources = append(sources, t)
				}
			}
			victim, ok := ctx.ClauseTarget(1)
			if !ok || victim.Kind != game.TargetCard || !ctx.IsTargetLegal(victim) {
				return nil
			}
			return eachDealsDamageEqualToItsPower(ctx, sources, victim.ID)
		},
	})
}
