package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// voting_a_helpers.go — the pieces the twelve voting cards of #2928's
// first slice share, on ADR 0146's vote (CR 701.38). Each card file
// keeps what is printed on it; what several of them need is here once.

// voteForThisCreature is the body of "When this creature enters,
// starting with you, each player votes for …" for a card whose result
// acts on the creature itself (Lieutenants of the Guard, Orchard
// Elemental, Messenger Jays, Council Guardian).
//
// "This creature" is the object that entered: one that has left, or
// come back as a new object, while the trigger waited is not acted on
// (CR 400.7). That is read as the trigger resolves and carried on the
// vote, since nothing can move the creature while the players vote —
// the shape Magister of Worth and Coercive Portal use.
func voteForThisCreature(v Vote) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if onBattlefield(g, item.SourceCardID) && !sourceIsNewObject(g, item) {
			v.Carry = []uuid.UUID{item.SourceCardID}
		}
		return v.Apply(NewContext(g, item))
	}
}

// votedCreature is the creature a voteForThisCreature vote was started
// for, or uuid.Nil when it had already left.
func votedCreature(r game.VoteResult) uuid.UUID {
	if len(r.Carry) == 0 {
		return uuid.Nil
	}
	return r.Carry[0]
}

// countersOnVotedCreature is "put N +1/+1 counters on this creature"
// as a vote's result.
func countersOnVotedCreature(ctx *Context, r game.VoteResult, n int) error {
	id := votedCreature(r)
	if id == uuid.Nil || n <= 0 {
		return nil
	}
	return AddCounter{Target: id, Kind: "+1/+1", N: n}.Apply(ctx)
}

// eachOpponentDiscards is "each opponent discards N cards", each
// choosing their own. Nothing hangs off it.
func eachOpponentDiscards(ctx *Context, n int) {
	if n <= 0 {
		return
	}
	ctx.Game.EachPlayerDiscardsForEffect(ctx.Controller(), game.DiscardPrompt{Source: ctx.Source(), N: n})
}
