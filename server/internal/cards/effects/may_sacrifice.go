package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// may_sacrifice.go — "you may sacrifice [a/another X]. If you do, …" as an
// ability resolves (CR 608.2c, 701.21).
//
// The old shape for these cards chose the permanent as the trigger's
// TARGET, so the pick was made as the trigger went on the stack: an
// opponent could answer it, hexproof and shroud made a creature
// unpickable, and a creature removed in response fizzled the whole
// trigger where the printed card lets you pick another. This asks at
// resolution instead, over whatever the controller controls THEN, with a
// floor of zero for the "may".
//
// `then` is told which permanent went. It runs only for a sacrifice that
// was chosen — a commander that takes the command zone instead (CR 903.9)
// still counts, because the engine hands the continuation what was
// chosen after the move has settled. Read the sacrificed creature's
// power with departedCreaturePower, which is its last-known information
// (CR 608.2h), so an anthem's bonus counts.

// maySacrificeOne queues the "you may sacrifice" prompt for the item's
// controller over the permanents `match` admits (the source is excluded
// when `another` is set). Nothing to sacrifice means no question.
func maySacrificeOne(g *game.Game, item *game.StackItem, question string, match CardPredicate, another bool, then func(ctx *Context, sacrificed uuid.UUID) error) error {
	return sacrificeOneThen(g, item, question, match, another, 0, then)
}

// mustSacrificeOne is maySacrificeOne without the "may": "sacrifice a
// non-Demon creature. If you do, …" (Dreadfeast Demon). The prompt has
// a floor of one, and a controller with nothing that matches is not
// asked and `then` does not run.
func mustSacrificeOne(g *game.Game, item *game.StackItem, question string, match CardPredicate, then func(ctx *Context, sacrificed uuid.UUID) error) error {
	return sacrificeOneThen(g, item, question, match, false, 1, then)
}

// sacrificeOneThen is the shared body: a prompt over the controller's
// matching permanents with the given floor and a ceiling of one.
func sacrificeOneThen(g *game.Game, item *game.StackItem, question string, match CardPredicate, another bool, floor int, then func(ctx *Context, sacrificed uuid.UUID) error) error {
	self := item.SourceCardID
	return ChoosePermanents{
		Question: question,
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			var out []uuid.UUID
			for _, id := range permanentsControlledByMatching(g, of, match) {
				if another && id == self {
					continue
				}
				out = append(out, id)
			}
			if len(out) == 0 {
				return nil, 0, 0
			}
			return out, floor, 1
		},
		Then: func(ctx *Context, picked game.PromptedPicks) error {
			cards := picked.Cards()
			if len(cards) == 0 {
				return nil
			}
			// Sacrificed here rather than by ChoosePermanents, because
			// "if you do" is about the permanent that REALLY left: a
			// CR 614 window that keeps it on the battlefield cancels the
			// sacrifice, and a commander that goes to the command zone
			// afterwards was still sacrificed (CR 701.17a, ADR 0115).
			return ctx.Game.SacrificeAllThenForEffect(uuid.Nil, cards, func(g *game.Game, sacrificed []uuid.UUID) error {
				if len(sacrificed) == 0 {
					return nil
				}
				return then(NewContext(g, ctx.Item), sacrificed[0])
			})
		},
	}.Apply(NewContext(g, item))
}
