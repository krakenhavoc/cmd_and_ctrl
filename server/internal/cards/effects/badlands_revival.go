package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Badlands Revival — Sorcery {3}{B}{G} (Outlaws of Thunder Junction,
// #4468):
//
//	"Return up to one target creature card from your graveyard to the
//	 battlefield. Return up to one target permanent card from your
//	 graveyard to your hand."
//
// Two clauses with different predicates, each "up to one" (Fates'
// Reversal's clause, twice), read back by slot. The two instructions are
// two steps (ResolveSteps), so the hand return waits for any prompt the
// creature's entry asks. Each half does only what its own still-legal
// target allows (CR 608.2b): the same creature card may be chosen for
// both clauses (CR 115.3), and then it goes to the battlefield and the
// second half finds a new object and does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a61267d2-2cc7-4208-a365-581d56b92e1c",
		Name:         "Badlands Revival",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCardInGraveyard("up to one target creature card from your graveyard", YouOwn(), Creature()).WithCount(0, 1),
			TargetCardInGraveyard("up to one target permanent card from your graveyard", YouOwn(), Permanent()).WithCount(0, 1),
		),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				t, ok := ctx.ClauseTarget(0)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				return ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
				t, ok := ctx.ClauseTarget(1)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				return ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}.Apply(ctx)
			}),
		),
	})
}
