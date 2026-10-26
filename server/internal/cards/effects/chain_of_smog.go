package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chain of Smog — Sorcery {1}{B}:
//
//	"Target player discards two cards. That player may copy this spell
//	 and may choose a new target for that copy."
//
// The second card of the Chain cycle, and the same CR 707.10
// self-copy as Chain of Vapor — see chain_of_vapor.go and
// game/resolving_item.go for why a resolving spell can find itself at
// all. What differs is only which player is asked and what happens
// before the question: the copy clause here is NOT conditional on the
// discard, so a player with an empty hand discards nothing and may
// still copy.
//
// The copy is controlled by the player who made it (CR 707.10b), which
// is what lets the chain walk around the table.
//
// The discarding player chooses the two cards (CR 701.9a: the player
// discarding picks, unless the text says "at random"), through the same
// discard prompt Thirst for Knowledge uses; the copy offer follows once
// the cards are in the graveyard, and also for a hand too small to ask
// about, because the copy clause is not conditional on the discard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ea14c26b-bf2f-48b4-b879-6e63069ded1f",
		Name:         "Chain of Smog",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return chainOfSmogDiscard(ctx)
		},
	})
}

// chainOfSmogDiscard makes the target player discard, then offers them
// the copy.
func chainOfSmogDiscard(ctx *Context) error {
	t, ok := ctx.ClauseTarget(0)
	if !ok || !ctx.IsTargetLegal(t) {
		return nil
	}
	player := t.ID
	if t.Kind == game.TargetSelf {
		player = ctx.Controller()
	} else if t.Kind != game.TargetPlayer {
		return nil
	}
	// The spell rides the discard as data (its ID and its controller): the
	// copy offer that follows names it, and the prompt holds no closure, so
	// a table waiting on the discard is a restore point (#4253).
	return ctx.Game.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
		Player:   player,
		Source:   ctx.Source(),
		N:        2,
		Question: "Chain of Smog — discard two cards",
	}, chainOfSmogDiscarded, []uuid.UUID{ctx.Item.ID, ctx.Controller()})
}

// chainOfSmogDiscarded offers the copy once the two cards are in the
// graveyard, and also for a hand too small to ask about: the copy clause is
// not conditional on the discard. Keyed (#4253); an on-disk key.
var chainOfSmogDiscarded = game.RegisterDiscardThen("choose-cards/chain-of-smog-discarded", func(g *game.Game, r game.DiscardsDone) error {
	if len(r.Carry) < 2 {
		return nil
	}
	ctx := keyedContext(g, r.Carry[1], r.Source)
	ctx.Item.ID = r.Carry[0]
	ctx.Item.Kind = game.StackItemSpell
	return MayChoice{
		Player:   r.Player,
		Question: "Chain of Smog — copy it? (you may choose a new target for the copy)",
		Then:     mayCopyThisSpell,
	}.Apply(ctx)
})
