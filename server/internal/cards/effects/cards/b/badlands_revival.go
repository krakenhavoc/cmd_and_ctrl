package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "a61267d2-2cc7-4208-a365-581d56b92e1c",
		Name:         "Badlands Revival",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetCardInGraveyard("up to one target creature card from your graveyard", effects.YouOwn(), effects.Creature()).WithCount(0, 1),
			effects.TargetCardInGraveyard("up to one target permanent card from your graveyard", effects.YouOwn(), effects.Permanent()).WithCount(0, 1),
		),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.ReturnClauseCardFromGraveyard(ctx, 0, game.ZoneBattlefield)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.ReturnClauseCardFromGraveyard(ctx, 1, game.ZoneHand)
			}),
		),
	})
}
