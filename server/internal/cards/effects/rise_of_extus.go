package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rise of Extus — Sorcery {4}{W/B}{W/B} (EDHREC rank 29151):
//
//	"Exile target creature. Exile up to one target instant or sorcery
//	 card from a graveyard.
//	 Learn. (You may reveal a Lesson card you own from outside the game
//	 and put it into your hand, or discard a card to draw a card.)"
//
// Two clauses, each with its own target and judged on its own at
// resolution (CR 608.2b), then learn (CR 701.48a, learn.go). Either
// exile can pause on a replacement's question, so each is its own step.
// The graveyard clause is "up to one", so the spell can be cast with
// only the creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "810f2b50-59c7-490a-8e6a-00f982cd16a0",
		Name:         "Rise of Extus",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCreature("target creature"),
			TargetCardInGraveyard("up to one target instant or sorcery card from a graveyard",
				Or(Instant(), Sorcery())).WithCount(0, 1),
		),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error { return exileClauseTarget(ctx, 0) }),
			Instruction(func(ctx *Context) error { return exileClauseTarget(ctx, 1) }),
			Learn{},
		),
	})
}

// exileClauseTarget exiles the pick that answered clause `slot`, if it is
// still legal as the spell resolves (CR 608.2b).
func exileClauseTarget(ctx *Context, slot int) error {
	t, ok := ctx.ClauseTarget(slot)
	if !ok || t.Kind != game.TargetCard || t.ID == uuid.Nil {
		return nil
	}
	return ExileTarget{Target: t.ID}.Apply(ctx)
}
