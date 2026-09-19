package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
// Declared simplification: the two cards discarded are chosen at
// random rather than by the discarding player. That is
// `DiscardCards`' behaviour everywhere in the catalog, not something
// this card introduces, and it is the reason the entry is not
// `CompletenessFull`.
func init() {
	Register(Spec{
		OracleID:     "ea14c26b-bf2f-48b4-b879-6e63069ded1f",
		Name:         "Chain of Smog",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The two cards discarded are picked at random rather than chosen by the player discarding them.",
		},
		Targets: TargetPlayer("target player"),
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
	if err := (DiscardCards{Player: player, N: 2}).Apply(ctx); err != nil {
		return err
	}
	return MayChoice{
		Player:   player,
		Question: "Chain of Smog — copy it? (you may choose a new target for the copy)",
		OnYes:    copyThisSpellFor(player),
	}.Apply(ctx)
}
