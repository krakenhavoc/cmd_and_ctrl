package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mob Verdict — Sorcery {2}{R}{R}:
//
//	"Secret council — Each player secretly votes for another player,
//	 then those votes are revealed. For each vote an opponent received,
//	 Mob Verdict deals 2 damage to that player and each creature that
//	 player controls. For each vote you received, draw a card."
//
// A secret vote over players (ADR 0180) in which no voter is offered
// themselves ("another player", Vote.OthersOnly). Each vote an opponent
// received is its own instruction: 2 damage to that player and each
// creature they control at once, one damage instance, and the next
// vote's damage after it has settled, so a creature that died to the
// first vote is not hit by the second. Opponents are dealt with in turn
// order, then the caster draws a card per vote they received.
//
// The bots' hints: the caster wants every vote on an opponent; an
// opponent wants the votes on another opponent rather than on the
// caster, whom a vote only gives a card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cb91e09e-f7fb-44bc-b78f-de29705bd2d6",
		Name:         "Mob Verdict",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			opts := VoteForPlayers(ctx.Game)
			forController := make([]int, len(opts))
			forOpponents := make([]int, len(opts))
			for i, o := range opts {
				if o.Player != ctx.Controller() {
					forController[i], forOpponents[i] = 2, 1
				}
			}
			return Vote{
				Question:      "Mob Verdict — secretly vote for another player",
				Options:       opts,
				Secret:        true,
				OthersOnly:    true,
				ForController: forController,
				ForOpponents:  forOpponents,
				Then:          mobVerdictVoted,
			}.Apply(ctx)
		},
	})
}

var mobVerdictVoted = VoteResultThen("vote/mob-verdict", func(ctx *Context, r game.VoteResult) error {
	received := VotesReceived(r)
	var hits []uuid.UUID
	for _, p := range ctx.Opponents() {
		for range received[p] {
			hits = append(hits, p)
		}
	}
	return mobVerdictStrike(ctx.Game, ctx.Source(), ctx.Controller(), hits, received[ctx.Controller()])
})

// mobVerdictStrike deals one vote's damage, 2 to the first player in
// hits and to each creature they control, and the rest once it has
// settled; then the caster draws `draws` cards. Captures only values.
func mobVerdictStrike(g *game.Game, source, caster uuid.UUID, hits []uuid.UUID, draws int) error {
	if len(hits) == 0 {
		if draws > 0 && voterStillIn(g, caster) {
			return g.DrawNForEffect(caster, draws)
		}
		return nil
	}
	p, rest := hits[0], hits[1:]
	if !voterStillIn(g, p) {
		return mobVerdictStrike(g, source, caster, rest, draws)
	}
	targets := []uuid.UUID{p}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == p && c.IsCreature() {
			targets = append(targets, c.InstanceID)
		}
	}
	return g.DealDamageEachThenForEffect(source, targets, 2, func(g *game.Game, _ int) error {
		return mobVerdictStrike(g, source, caster, rest, draws)
	})
}
