package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Trap the Trespassers — Instant {2}{U}:
//
//	"Secret council — Each player secretly votes for a creature you
//	 don't control, then those votes are revealed. For each creature
//	 with one or more votes, put that many stun counters on it, then
//	 tap it."
//
// A secret object vote (ADR 0180) over the creatures the caster doesn't
// control as the spell resolves. Nothing is targeted. Each creature that
// got votes and is still on the battlefield gets one stun counter per
// vote and is then tapped, creature by creature in the order the vote
// listed them. With no such creature nobody votes.
//
// A bot votes for the best creature it doesn't control and never for
// its own; it cannot see how the others voted, so it does not lean.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ef72af33-37fc-48ed-90d8-547c90e1609d",
		Name:         "Trap the Trespassers",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			var ids []uuid.UUID
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.Controller != ctx.Controller() && c.IsCreature() {
					ids = append(ids, c.InstanceID)
				}
			}
			return Vote{
				Question: "Trap the Trespassers — secretly vote for a creature its caster doesn't control",
				Options:  VoteForPermanents(ctx.Game, ids),
				Secret:   true,
				Then:     trapTheTrespassersVoted,
			}.Apply(ctx)
		},
	})
}

var trapTheTrespassersVoted = VoteResultThen("vote/trap-the-trespassers", func(ctx *Context, r game.VoteResult) error {
	for i := range r.Options {
		n := r.Votes(i)
		id := VotedPermanent(r, i)
		if n == 0 || id == uuid.Nil || !onBattlefield(ctx.Game, id) {
			continue
		}
		if err := (AddCounter{Target: id, Kind: game.CounterStun, N: n}).Apply(ctx); err != nil {
			return err
		}
		if err := (TapTarget{Target: id}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
