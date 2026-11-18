package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Duel for Dominance — Instant {1}{G} (Midnight Hunt, #5497):
//
//	"Coven — Choose target creature you control and target creature you
//	 don't control. If you control three or more creatures with different
//	 powers, put a +1/+1 counter on the chosen creature you control. Then
//	 the chosen creatures fight each other."
//
// Longstalk Brawl's shape: the counter (when coven holds as the spell
// resolves) goes first, then the fight between the still-legal picks. A
// pick that is no longer legal gets no counter and no fight (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5064a55a-9692-4a67-9e82-16250b9140d1",
		Name:         "Duel for Dominance",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			TargetCreature("target creature you don't control", OpponentControls()),
		),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				mine, ok := ctx.ClauseTarget(0)
				if !ok || mine.Kind != game.TargetCard || !trekB48Coven(ctx.Game, ctx.Controller(), ctx.Item.SourceCardID) {
					return nil
				}
				return AddCounter{Target: mine.ID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
			}),
			Instruction(fightTargetClauses),
		),
	})
}
