package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Make an Example — Sorcery {3}{B} (Streets of New Capenna, #5491):
//
//	"Each opponent separates the creatures they control into two piles. For
//	 each opponent, you choose one of their piles. Each opponent sacrifices the
//	 creatures in their chosen pile. (Piles can be empty.)"
//
// Do or Die turned around: each OPPONENT is the splitter (a reveal-pick prompt
// addressed to them, an empty pile being legal) and YOU choose which pile goes.
// The opponents are walked one at a time, in seat order, and every pile is made
// and chosen BEFORE anything is sacrificed: the piles are cut from the board as
// it stood when the spell resolved, and the creatures of every chosen pile are
// sacrificed together, one simultaneous event (CR 701.21a), so a sacrifice
// trigger can't change what a later opponent splits. The walk is carried in the
// keyed tail (the opponents still to split, then the creatures already chosen),
// so a table waiting on either prompt is a restore point. An opponent with no
// creatures asks nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID: "591edddc-3736-44cf-9719-117c24ff3980",
		Name:     "Make an Example",
		// ADR 0126 §6: one pile of each opponent's creatures.
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepSacrifice, OpponentsOnly: true, Partial: true}},
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return makeAnExampleNext(ctx, ctx.Opponents(), nil)
		},
	})
}

// makeAnExampleNext starts the next opponent's split, or sacrifices every
// creature in the piles chosen so far once none is left to split.
func makeAnExampleNext(ctx *Context, opponents, chosen []uuid.UUID) error {
	for i, opp := range opponents {
		cards := creaturesControlledByPlayer(ctx.Game, opp)
		if len(cards) == 0 {
			continue
		}
		rest := opponents[i+1:]
		tail := ctx.Tail(makeAnExampleChosen)
		tail.N = len(rest)
		tail.Carry = append(append([]uuid.UUID(nil), rest...), chosen...)
		return PileSplit{
			Splitter:      opp,
			Chooser:       ctx.Controller(),
			Owner:         opp,
			SplitQuestion: "Make an Example — separate the creatures you control into two piles",
			PickQuestion:  "Make an Example — choose a pile; its creatures are sacrificed",
			Cards:         cards,
			ThenTail:      tail,
		}.Apply(ctx)
	}
	if len(chosen) > 0 {
		ctx.Game.SacrificeAllForEffect(ctx.Source(), chosen)
	}
	return nil
}

// makeAnExampleChosen is the tail after each opponent's pile is chosen: it adds
// the chosen pile ("taken") to the creatures to sacrifice and moves on. Keyed
// (#4253): both prompts carry it. An on-disk key: never renamed, never reused.
var makeAnExampleChosen game.TailRef

// The tail is registered from init because it and makeAnExampleNext call each
// other, which a package-level initialiser may not.
func init() {
	makeAnExampleChosen = PileTail("then/make-an-example-chosen-pile", func(ctx *Context, taken, _ []uuid.UUID, o game.TailOutcome) error {
		n := min(max(o.Tail.N, 0), len(o.Tail.Carry))
		rest, chosen := o.Tail.Carry[:n:n], o.Tail.Carry[n:]
		return makeAnExampleNext(ctx, rest, append(append([]uuid.UUID(nil), chosen...), taken...))
	})
}
